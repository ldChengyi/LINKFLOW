class Device {
  final String id;
  final String name;
  final String? modelId;
  final String? modelName;
  final String status;

  Device({
    required this.id,
    required this.name,
    this.modelId,
    this.modelName,
    required this.status,
  });

  factory Device.fromJson(Map<String, dynamic> json) {
    return Device(
      id: json['id'],
      name: json['name'],
      modelId: json['model_id'],
      modelName: json['model_name'],
      status: json['status'] ?? 'offline',
    );
  }

  bool get isOnline => status == 'online';
}
