import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';
import '../services/ws_service.dart';
import '../models/device.dart';
import 'chat_screen.dart';

class DeviceListScreen extends StatefulWidget {
  const DeviceListScreen({super.key});

  @override
  State<DeviceListScreen> createState() => _DeviceListScreenState();
}

class _DeviceListScreenState extends State<DeviceListScreen> {
  List<Device> _devices = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadDevices();
  }

  Future<void> _loadDevices() async {
    setState(() => _loading = true);
    try {
      final token = context.read<AuthService>().token;
      final response = await ApiService.request('GET', '/api/devices', token: token);
      setState(() {
        _devices = (response['data']['list'] as List).map((d) => Device.fromJson(d)).toList();
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    // 监听 WSService 的设备状态变化
    final wsService = context.watch<WSService>();

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: const Color(0xFF2196F3),
        elevation: 0,
        title: const Text('我的设备', style: TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFF2196F3)))
          : RefreshIndicator(
              color: const Color(0xFF2196F3),
              onRefresh: _loadDevices,
              child: _devices.isEmpty
                  ? ListView(children: const [
                      SizedBox(height: 200),
                      Center(child: Text('暂无设备', style: TextStyle(color: Color(0xFFBDBDBD), fontSize: 16))),
                    ])
                  : ListView.builder(
                      padding: const EdgeInsets.all(12),
                      itemCount: _devices.length,
                      itemBuilder: (ctx, i) {
                        final device = _devices[i];
                        // WS 推送的状态优先，否则用 API 返回的
                        final isOnline = wsService.deviceOnline.containsKey(device.id)
                            ? wsService.isOnline(device.id)
                            : device.isOnline;
                        return _DeviceItem(
                          device: device,
                          isOnline: isOnline,
                          index: i,
                          onTap: () {
                            Navigator.push(context, MaterialPageRoute(
                              builder: (_) => ChatScreen(device: device),
                            ));
                          },
                        );
                      },
                    ),
            ),
    );
  }
}

class _DeviceItem extends StatefulWidget {
  final Device device;
  final bool isOnline;
  final int index;
  final VoidCallback onTap;

  const _DeviceItem({required this.device, required this.isOnline, required this.index, required this.onTap});

  @override
  State<_DeviceItem> createState() => _DeviceItemState();
}

class _DeviceItemState extends State<_DeviceItem> with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _fadeAnim;
  late Animation<Offset> _slideAnim;
  bool _pressed = false;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: const Duration(milliseconds: 400));
    _fadeAnim = CurvedAnimation(parent: _controller, curve: Curves.easeOut);
    _slideAnim = Tween(begin: const Offset(0, 0.15), end: Offset.zero)
        .animate(CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic));
    Future.delayed(Duration(milliseconds: widget.index * 60), () {
      if (mounted) _controller.forward();
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _fadeAnim,
      child: SlideTransition(
        position: _slideAnim,
        child: GestureDetector(
          onTapDown: (_) => setState(() => _pressed = true),
          onTapUp: (_) {
            setState(() => _pressed = false);
            widget.onTap();
          },
          onTapCancel: () => setState(() => _pressed = false),
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            margin: const EdgeInsets.only(bottom: 8),
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: _pressed ? const Color(0xFFE3F2FD) : Colors.white,
              borderRadius: BorderRadius.circular(14),
              boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 8, offset: const Offset(0, 2))],
            ),
            child: Row(
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: const Color(0xFFE3F2FD),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(Icons.developer_board, color: Color(0xFF2196F3), size: 26),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(widget.device.name, style: const TextStyle(color: Color(0xFF212121), fontSize: 16, fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      Text(widget.device.modelName ?? '未绑定物模型', style: const TextStyle(color: Color(0xFF757575), fontSize: 13)),
                    ],
                  ),
                ),
                AnimatedContainer(
                  duration: const Duration(milliseconds: 300),
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: widget.isOnline ? const Color(0xFFE8F5E9) : const Color(0xFFF5F5F5),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      AnimatedContainer(
                        duration: const Duration(milliseconds: 300),
                        width: 8,
                        height: 8,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: widget.isOnline ? const Color(0xFF4CAF50) : const Color(0xFFBDBDBD),
                        ),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        widget.isOnline ? '在线' : '离线',
                        style: TextStyle(
                          color: widget.isOnline ? const Color(0xFF4CAF50) : const Color(0xFF757575),
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
