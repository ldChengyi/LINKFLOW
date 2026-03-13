import 'package:flutter/material.dart';
import 'package:speech_to_text/speech_to_text.dart';

class VoiceInputBar extends StatefulWidget {
  final Function(String text, {bool isVoice}) onSend;

  const VoiceInputBar({super.key, required this.onSend});

  @override
  State<VoiceInputBar> createState() => _VoiceInputBarState();
}

class _VoiceInputBarState extends State<VoiceInputBar> with SingleTickerProviderStateMixin {
  final _controller = TextEditingController();
  final _speech = SpeechToText();
  bool _isListening = false;
  late AnimationController _pulseController;
  late Animation<double> _pulseAnim;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(vsync: this, duration: const Duration(milliseconds: 800));
    _pulseAnim = Tween(begin: 1.0, end: 1.3).animate(CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut));
  }

  @override
  void dispose() {
    _controller.dispose();
    _pulseController.dispose();
    super.dispose();
  }

  Future<void> _startListening() async {
    if (!await _speech.initialize(onError: (e) => debugPrint('Speech error: $e'))) return;
    setState(() => _isListening = true);
    _pulseController.repeat(reverse: true);
    _speech.listen(onResult: (result) {
      if (result.finalResult) {
        widget.onSend(result.recognizedWords, isVoice: true);
        setState(() => _isListening = false);
        _pulseController.stop();
        _pulseController.reset();
      }
    }, localeId: 'zh_CN');
  }

  void _stopListening() {
    _speech.stop();
    setState(() => _isListening = false);
    _pulseController.stop();
    _pulseController.reset();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 10, offset: co2请22nst Offset(0, -2))],
      ),
      child: SafeArea(
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _controller,
                style: const TextStyle(color: Color(0xFF212121)),
                decoration: InputDecoration(
                  hintText: '输入消息...',
                  hintStyle: const TextStyle(color: Color(0xFFBDBDBD)),
                  filled: true,
                  fillColor: const Color(0xFFF5F5F5),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(24), borderSide: BorderSide.none),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                ),
              ),
            ),
            const SizedBox(width: 8),
            // 语音按钮
            GestureDetector(
              behavior: HitTestBehavior.opaque,
              onLongPressStart: (_) => _startListening(),
              onLongPressEnd: (_) => _stopListening(),
              child: ScaleTransition(
                scale: _isListening ? _pulseAnim : const AlwaysStoppedAnimation(1.0),
                child: Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: _isListening ? const Color(0xFFE53935) : const Color(0xFF2196F3),
                    boxShadow: [
                      BoxShadow(
                        color: (_isListening ? const Color(0xFFE53935) : const Color(0xFF2196F3)).withOpacity(0.3),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Icon(_isListening ? Icons.mic : Icons.mic_none, color: Colors.white, size: 22),
                ),
              ),
            ),
            const SizedBox(width: 8),
            // 发送按钮
            Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: () {
                  if (_controller.text.trim().isNotEmpty) {
                    widget.onSend(_controller.text.trim());
                    _controller.clear();
                  }
                },
                customBorder: const CircleBorder(),
                child: Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: const Color(0xFF2196F3),
                    boxShadow: [BoxShadow(color: const Color(0xFF2196F3).withOpacity(0.3), blurRadius: 8, offset: const Offset(0, 2))],
                  ),
                  child: const Icon(Icons.send, color: Colors.white, size: 20),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
