import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/auth_service.dart';
import '../services/api_service.dart';
import '../services/ws_service.dart';
import '../models/device.dart';
import '../models/chat_message.dart';
import '../widgets/chat_bubble.dart';
import '../widgets/voice_input_bar.dart';
import '../services/chat_db.dart';
import '../config.dart';
import 'device_data_screen.dart';

class ChatScreen extends StatefulWidget {
  final Device device;

  const ChatScreen({super.key, required this.device});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final List<ChatMessage> _messages = [];

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    final history = await ChatDB.getMessages(widget.device.id);
    if (mounted) setState(() => _messages.addAll(history));
  }

  void _addSystemMessage(String content) {
    final msg = ChatMessage(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: MessageType.systemInfo,
      content: content,
      timestamp: DateTime.now(),
    );
    setState(() => _messages.insert(0, msg));
    ChatDB.insertMessage(widget.device.id, msg);
  }

  /// 将相对 audio_url 转为完整 URL
  String? _buildAudioUrl(String? audioUrl) {
    if (audioUrl == null || audioUrl.isEmpty) return null;
    if (audioUrl.startsWith('http')) return audioUrl;
    final base = ApiService.baseUrl;
    if (base == null) return null;
    return '$base$audioUrl';
  }

  Future<void> _sendMessage(String text, {bool isVoice = false}) async {
    final userMsg = ChatMessage(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: isVoice ? MessageType.userVoice : MessageType.userText,
      content: text,
      timestamp: DateTime.now(),
    );
    setState(() => _messages.insert(0, userMsg));
    ChatDB.insertMessage(widget.device.id, userMsg);

    try {
      final token = context.read<AuthService>().token;
      final response = await ApiService.request(
        'POST',
        '/api/devices/${widget.device.id}/voice-debug',
        body: {'text': text},
        token: token,
        timeout: AppConfig.voiceTimeout,
      );
      final result = response['data'];
      final fullAudioUrl = _buildAudioUrl(result['audio_url']);
      debugPrint('🔊 voice-debug response audio_url: ${result['audio_url']} → $fullAudioUrl');
      final replyMsg = ChatMessage(
        id: DateTime.now().millisecondsSinceEpoch.toString(),
        type: MessageType.deviceReply,
        content: result['message'],
        success: result['success'],
        action: result['action'],
        audioUrl: fullAudioUrl,
        timestamp: DateTime.now(),
      );
      setState(() => _messages.insert(0, replyMsg));
      ChatDB.insertMessage(widget.device.id, replyMsg);
    } catch (e) {
      _addSystemMessage('发送失败: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    final ws = context.watch<WSService>();
    final isOnline = ws.deviceOnline.containsKey(widget.device.id)
        ? ws.isOnline(widget.device.id)
        : widget.device.isOnline;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: const Color(0xFF2196F3),
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
        title: Row(
          children: [
            Text(widget.device.name, style: const TextStyle(color: Colors.white)),
            const SizedBox(width: 8),
            Container(
              width: 10,
              height: 10,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: isOnline ? const Color(0xFF4CAF50) : const Color(0xFFBDBDBD),
                boxShadow: isOnline
                    ? [BoxShadow(color: const Color(0xFF4CAF50).withOpacity(0.5), blurRadius: 6)]
                    : null,
              ),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.bar_chart_rounded, color: Colors.white),
            tooltip: '设备数据',
            onPressed: () {
              Navigator.push(context, MaterialPageRoute(
                builder: (_) => DeviceDataScreen(device: widget.device),
              ));
            },
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              reverse: true,
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length,
              itemBuilder: (ctx, i) {
                final msg = _messages[i];
                // 时间分隔：与下一条（更早的）消息比较
                final prevMsg = (i + 1 < _messages.length) ? _messages[i + 1] : null;
                final showTime = shouldShowTimeSeparator(
                  prevMsg?.timestamp,
                  msg.timestamp,
                );
                return Column(
                  children: [
                    if (showTime) ChatTimeSeparator(time: msg.timestamp),
                    ChatBubble(message: msg),
                  ],
                );
              },
            ),
          ),
          VoiceInputBar(onSend: _sendMessage),
        ],
      ),
    );
  }
}
