enum MessageType {
  userVoice,
  userText,
  deviceReply,
  deviceData,
  systemInfo,
}

class ChatMessage {
  final String id;
  final MessageType type;
  final String content;
  final String? audioUrl;
  final String? action;
  final bool? success;
  final Map<String, dynamic>? data;
  final DateTime timestamp;

  ChatMessage({
    required this.id,
    required this.type,
    required this.content,
    this.audioUrl,
    this.action,
    this.success,
    this.data,
    required this.timestamp,
  });

  Map<String, dynamic> toMap(String deviceId) => {
    'id': id,
    'device_id': deviceId,
    'type': type.name,
    'content': content,
    'audio_url': audioUrl,
    'action': action,
    'success': success == null ? null : (success! ? 1 : 0),
    'data': data?.toString(),
    'timestamp': timestamp.millisecondsSinceEpoch,
  };

  factory ChatMessage.fromMap(Map<String, dynamic> map) => ChatMessage(
    id: map['id'],
    type: MessageType.values.firstWhere((e) => e.name == map['type']),
    content: map['content'],
    audioUrl: map['audio_url'],
    action: map['action'],
    success: map['success'] == null ? null : map['success'] == 1,
    data: map['data'] != null ? {'raw': map['data']} : null,
    timestamp: DateTime.fromMillisecondsSinceEpoch(map['timestamp']),
  );
}
