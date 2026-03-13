import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';
import '../services/ws_service.dart';
import '../models/device.dart';

class DeviceDataScreen extends StatefulWidget {
  final Device device;

  const DeviceDataScreen({super.key, required this.device});

  @override
  State<DeviceDataScreen> createState() => _DeviceDataScreenState();
}

class _DeviceDataScreenState extends State<DeviceDataScreen> {
  Map<String, dynamic> _data = {};
  Map<String, PropertyMeta> _propertyMeta = {};
  bool _loading = true;
  DateTime? _lastUpdate;
  StreamSubscription? _wsSub;

  @override
  void initState() {
    super.initState();
    _loadThingModel();
    _loadLatest();
    // 监听 WS 遥测消息
    final ws = context.read<WSService>();
    _wsSub = ws.messages.listen((msg) {
      if (msg.type == 'telemetry' && msg.data is Map<String, dynamic>) {
        final data = msg.data as Map<String, dynamic>;
        if (data['device_id'] == widget.device.id && data['payload'] is Map<String, dynamic>) {
          if (mounted) {
            setState(() {
              _data = {..._data, ...(data['payload'] as Map<String, dynamic>)};
              _lastUpdate = DateTime.now();
            });
          }
        }
      }
    });
  }

  Future<void> _loadThingModel() async {
    if (widget.device.modelId == null) return;
    try {
      final token = context.read<AuthService>().token;
      final response = await ApiService.request(
        'GET',
        '/api/thing-models/${widget.device.modelId}',
        token: token,
      );
      final model = response['data'];
      final props = model['properties'] as List?;
      if (props != null) {
        final meta = <String, PropertyMeta>{};
        for (var p in props) {
          meta[p['id']] = PropertyMeta(
            name: p['name'] ?? p['id'],
            unit: p['unit'] ?? '',
          );
        }
        setState(() => _propertyMeta = meta);
      }
    } catch (_) {}
  }

  Future<void> _loadLatest() async {
    try {
      final token = context.read<AuthService>().token;
      final response = await ApiService.request(
        'GET',
        '/api/devices/${widget.device.id}/data/latest',
        token: token,
      );
      final result = response['data'];
      if (result != null && result['payload'] is Map<String, dynamic>) {
        setState(() {
          _data = result['payload'] as Map<String, dynamic>;
          _lastUpdate = DateTime.now();
        });
      }
    } catch (_) {}
    setState(() => _loading = false);
  }

  @override
  void dispose() {
    _wsSub?.cancel();
    super.dispose();
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
        title: Text(widget.device.name, style: const TextStyle(color: Colors.white)),
        actions: [
          Container(
            margin: const EdgeInsets.only(right: 12),
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: isOnline ? Colors.white.withOpacity(0.2) : Colors.white.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: isOnline ? const Color(0xFF4CAF50) : const Color(0xFFBDBDBD),
                  ),
                ),
                const SizedBox(width: 4),
                Text(
                  isOnline ? '在线' : '离线',
                  style: const TextStyle(color: Colors.white, fontSize: 12),
                ),
              ],
            ),
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFF2196F3)))
          : RefreshIndicator(
              onRefresh: _loadLatest,
              color: const Color(0xFF2196F3),
              child: _data.isEmpty
                  ? ListView(children: const [
                      SizedBox(height: 200),
                      Center(child: Text('暂无数据', style: TextStyle(color: Color(0xFFBDBDBD), fontSize: 16))),
                    ])
                  : ListView(
                      padding: const EdgeInsets.all(16),
                      children: [
                        // 更新时间
                        if (_lastUpdate != null)
                          Padding(
                            padding: const EdgeInsets.only(bottom: 12),
                            child: Row(
                              children: [
                                const Icon(Icons.access_time, size: 14, color: Color(0xFF9E9E9E)),
                                const SizedBox(width: 4),
                                Text(
                                  '最后更新: ${_formatTime(_lastUpdate!)}',
                                  style: const TextStyle(color: Color(0xFF9E9E9E), fontSize: 12),
                                ),
                              ],
                            ),
                          ),
                        // 数据卡片
                        ..._data.entries.map((e) {
                          final meta = _propertyMeta[e.key];
                          return _DataTile(
                            id: e.key,
                            name: meta?.name ?? e.key,
                            unit: meta?.unit ?? '',
                            value: e.value,
                          );
                        }),
                      ],
                    ),
            ),
    );
  }

  String _formatTime(DateTime t) {
    return '${t.hour.toString().padLeft(2, '0')}:${t.minute.toString().padLeft(2, '0')}:${t.second.toString().padLeft(2, '0')}';
  }
}

class PropertyMeta {
  final String name;
  final String unit;
  PropertyMeta({required this.name, required this.unit});
}

class _DataTile extends StatelessWidget {
  final String id;
  final String name;
  final String unit;
  final dynamic value;

  const _DataTile({required this.id, required this.name, required this.unit, required this.value});

  @override
  Widget build(BuildContext context) {
    final isBool = value is bool || value == 0 || value == 1;
    final isNum = value is num;

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 8, offset: const Offset(0, 2))],
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: const Color(0xFFE3F2FD),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(
              isBool ? Icons.toggle_on_outlined : (isNum ? Icons.speed : Icons.text_fields),
              color: const Color(0xFF2196F3),
              size: 22,
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Text(name, style: const TextStyle(color: Color(0xFF757575), fontSize: 14)),
          ),
          Text(
            unit.isEmpty ? '$value' : '$value $unit',
            style: const TextStyle(color: Color(0xFF212121), fontSize: 16, fontWeight: FontWeight.w600),
          ),
        ],
      ),
    );
  }
}
