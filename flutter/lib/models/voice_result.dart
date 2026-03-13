class VoiceResult {
  final bool success;
  final String message;
  final String? action;
  final String? audioUrl;

  VoiceResult({
    required this.success,
    required this.message,
    this.action,
    this.audioUrl,
  });

  factory VoiceResult.fromJson(Map<String, dynamic> json) {
    return VoiceResult(
      success: json['success'] ?? false,
      message: json['message'] ?? '',
      action: json['action'],
      audioUrl: json['audio_url'],
    );
  }
}
