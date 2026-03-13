import 'package:flutter/material.dart';
import 'package:just_audio/just_audio.dart';
import '../models/chat_message.dart';

class ChatTimeSeparator extends StatelessWidget {
  final DateTime time;

  const ChatTimeSeparator({super.key, required this.time});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 10),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 3),
        decoration: BoxDecoration(
          color: const Color(0xFFE0E0E0),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Text(_formatTime(time), style: const TextStyle(color: Color(0xFF757575), fontSize: 11)),
      ),
    );
  }

  String _formatTime(DateTime t) {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final date = DateTime(t.year, t.month, t.day);
    final diff = today.difference(date).inDays;
    final timeStr = '${t.hour.toString().padLeft(2, '0')}:${t.minute.toString().padLeft(2, '0')}';

    if (diff == 0) return '今天 $timeStr';
    if (diff == 1) return '昨天 $timeStr';
    if (diff < 7) return '${_weekday(t.weekday)} $timeStr';
    return '${t.month}月${t.day}日 $timeStr';
  }

  String _weekday(int w) {
    const days = ['', '周一', '周二', '周三', '周四', '周五', '周六', '周日'];
    return days[w];
  }
}

/// 判断两条消息之间是否需要显示时间分隔
bool shouldShowTimeSeparator(DateTime? prev, DateTime current) {
  if (prev == null) return true;
  return current.difference(prev).inMinutes.abs() >= 5;
}

class ChatBubble extends StatefulWidget {
  final ChatMessage message;

  const ChatBubble({super.key, required this.message});

  @override
  State<ChatBubble> createState() => _ChatBubbleState();
}

class _ChatBubbleState extends State<ChatBubble> with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<Offset> _slideAnim;
  late Animation<double> _fadeAnim;
  AudioPlayer? _player;
  bool _playing = false;

  @override
  void initState() {
    super.initState();
    final isUser = widget.message.type == MessageType.userVoice || widget.message.type == MessageType.userText;
    _controller = AnimationController(vsync: this, duration: const Duration(milliseconds: 300));
    _slideAnim = Tween(
      begin: Offset(isUser ? 0.3 : -0.3, 0),
      end: Offset.zero,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic));
    _fadeAnim = CurvedAnimation(parent: _controller, curve: Curves.easeOut);
    _controller.forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    _player?.dispose();
    super.dispose();
  }

  Future<void> _toggleAudio() async {
    if (_playing) {
      await _player?.stop();
      setState(() => _playing = false);
      return;
    }
    _player ??= AudioPlayer();
    try {
      final url = widget.message.audioUrl!;
      debugPrint('🎵 Playing audio: $url');
      await _player!.setUrl(url);
      setState(() => _playing = true);
      _player!.play();
      _player!.playerStateStream.listen((state) {
        if (state.processingState == ProcessingState.completed) {
          if (mounted) setState(() => _playing = false);
        }
      });
    } catch (e) {
      debugPrint('❌ Audio playback error: $e');
      if (mounted) setState(() => _playing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final isUser = widget.message.type == MessageType.userVoice || widget.message.type == MessageType.userText;
    final isSystem = widget.message.type == MessageType.systemInfo;
    final hasAudio = widget.message.audioUrl != null && widget.message.audioUrl!.isNotEmpty;

    if (isSystem) {
      return FadeTransition(
        opacity: _fadeAnim,
        child: Center(
          child: Container(
            margin: const EdgeInsets.symmetric(vertical: 8),
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
            decoration: BoxDecoration(
              color: const Color(0xFFE0E0E0),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Text(widget.message.content, style: const TextStyle(color: Color(0xFF757575), fontSize: 12)),
          ),
        ),
      );
    }

    return SlideTransition(
      position: _slideAnim,
      child: FadeTransition(
        opacity: _fadeAnim,
        child: Align(
          alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
          child: Container(
            margin: const EdgeInsets.only(bottom: 12),
            padding: const EdgeInsets.all(12),
            constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
            decoration: BoxDecoration(
              color: isUser ? const Color(0xFF2196F3) : Colors.white,
              borderRadius: BorderRadius.only(
                topLeft: const Radius.circular(16),
                topRight: const Radius.circular(16),
                bottomLeft: Radius.circular(isUser ? 16 : 4),
                bottomRight: Radius.circular(isUser ? 4 : 16),
              ),
              boxShadow: [
                BoxShadow(color: Colors.black.withOpacity(0.06), blurRadius: 8, offset: const Offset(0, 2)),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  widget.message.content,
                  style: TextStyle(color: isUser ? Colors.white : const Color(0xFF212121), fontSize: 15),
                ),
                if (widget.message.action != null) ...[
                  const SizedBox(height: 4),
                  Text(widget.message.action!, style: TextStyle(color: isUser ? Colors.white70 : const Color(0xFF757575), fontSize: 11)),
                ],
                if (widget.message.success != null) ...[
                  const SizedBox(height: 4),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        widget.message.success! ? Icons.check_circle : Icons.error,
                        size: 14,
                        color: widget.message.success! ? const Color(0xFF4CAF50) : const Color(0xFFE53935),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        widget.message.success! ? '执行成功' : '执行失败',
                        style: TextStyle(fontSize: 11, color: widget.message.success! ? const Color(0xFF4CAF50) : const Color(0xFFE53935)),
                      ),
                    ],
                  ),
                ],
                // 语音播放按钮
                if (hasAudio && !isUser) ...[
                  const SizedBox(height: 6),
                  Material(
                    color: Colors.transparent,
                    child: InkWell(
                      onTap: _toggleAudio,
                      borderRadius: BorderRadius.circular(14),
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                        decoration: BoxDecoration(
                          color: const Color(0xFFF5F5F5),
                          borderRadius: BorderRadius.circular(14),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              _playing ? Icons.stop_circle_outlined : Icons.play_circle_outline,
                              size: 18,
                              color: const Color(0xFF2196F3),
                            ),
                            const SizedBox(width: 4),
                            Text(
                              _playing ? '停止' : '播放语音',
                              style: const TextStyle(color: Color(0xFF2196F3), fontSize: 12),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
