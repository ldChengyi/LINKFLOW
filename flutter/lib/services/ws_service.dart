import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

class WSService extends ChangeNotifier {
  WebSocketChannel? _channel;
  Timer? _reconnectTimer;
  int _reconnectDelay = 3;
  static const int _maxReconnectDelay = 30;
  String? _token;
  String? _baseUrl;
  bool _disposed = false;

  // 设备在线状态 map: deviceId → bool
  final Map<String, bool> _deviceOnline = {};
  // 最新遥测数据 map: deviceId → payload
  final Map<String, Map<String, dynamic>> _telemetryData = {};

  // 消息流，供页面监听
  final _messageController = StreamController<WSMessage>.broadcast();
  Stream<WSMessage> get messages => _messageController.stream;

  Map<String, bool> get deviceOnline => _deviceOnline;
  Map<String, Map<String, dynamic>> get telemetryData => _telemetryData;

  bool isOnline(String deviceId) => _deviceOnline[deviceId] ?? false;

  Map<String, dynamic>? getLatestData(String deviceId) => _telemetryData[deviceId];

  void connect(String baseUrl, String token) {
    _baseUrl = baseUrl;
    _token = token;
    _doConnect();
  }

  void _doConnect() {
    if (_disposed || _token == null || _baseUrl == null) return;

    disconnect(reconnect: false);

    // http → ws, https → wss
    final wsUrl = _baseUrl!
        .replaceFirst('http://', 'ws://')
        .replaceFirst('https://', 'wss://');
    final uri = Uri.parse('$wsUrl/api/ws?token=$_token');

    try {
      _channel = WebSocketChannel.connect(uri);
      _reconnectDelay = 3;

      _channel!.stream.listen(
        (data) {
          try {
            final json = jsonDecode(data as String) as Map<String, dynamic>;
            final msg = WSMessage.fromJson(json);
            _handleMessage(msg);
            _messageController.add(msg);
          } catch (_) {}
        },
        onDone: () => _scheduleReconnect(),
        onError: (_) => _scheduleReconnect(),
      );
    } catch (_) {
      _scheduleReconnect();
    }
  }

  void _handleMessage(WSMessage msg) {
    switch (msg.type) {
      case 'device_status':
        final data = msg.data;
        if (data is Map<String, dynamic>) {
          final deviceId = data['device_id'] as String?;
          final status = data['status'] as String?;
          if (deviceId != null && status != null) {
            _deviceOnline[deviceId] = status == 'online';
            notifyListeners();
          }
        }
        break;
      case 'telemetry':
        final data = msg.data;
        if (data is Map<String, dynamic>) {
          final deviceId = data['device_id'] as String?;
          final payload = data['payload'];
          if (deviceId != null && payload is Map<String, dynamic>) {
            _telemetryData[deviceId] = payload;
            notifyListeners();
          }
        }
        break;
      default:
        break;
    }
  }

  void _scheduleReconnect() {
    if (_disposed) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(Duration(seconds: _reconnectDelay), () {
      _reconnectDelay = (_reconnectDelay * 1.5).toInt().clamp(3, _maxReconnectDelay);
      _doConnect();
    });
  }

  void disconnect({bool reconnect = false}) {
    _reconnectTimer?.cancel();
    _channel?.sink.close();
    _channel = null;
    if (!reconnect) {
      _deviceOnline.clear();
      _telemetryData.clear();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    disconnect();
    _messageController.close();
    super.dispose();
  }
}

class WSMessage {
  final String type;
  final dynamic data;

  WSMessage({required this.type, required this.data});

  factory WSMessage.fromJson(Map<String, dynamic> json) {
    return WSMessage(type: json['type'] ?? '', data: json['data']);
  }
}
