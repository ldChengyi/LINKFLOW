import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';
import '../models/device.dart';

class DevicePicker extends StatefulWidget {
  final Function(Device) onSelect;

  const DevicePicker({super.key, required this.onSelect});

  @override
  State<DevicePicker> createState() => _DevicePickerState();
}

class _DevicePickerState extends State<DevicePicker> {
  List<Device> _devices = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadDevices();
  }

  Future<void> _loadDevices() async {
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
    return AlertDialog(
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      title: const Text('选择设备', style: TextStyle(color: Color(0xFF212121), fontWeight: FontWeight.w600)),
      content: _loading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFF2196F3)))
          : SizedBox(
              width: double.maxFinite,
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: _devices.length,
                itemBuilder: (ctx, i) {
                  final device = _devices[i];
                  return AnimatedContainer(
                    duration: Duration(milliseconds: 200 + i * 50),
                    curve: Curves.easeOut,
                    child: Card(
                      elevation: 0,
                      color: const Color(0xFFF5F5F5),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      margin: const EdgeInsets.only(bottom: 8),
                      child: ListTile(
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        title: Text(device.name, style: const TextStyle(color: Color(0xFF212121), fontWeight: FontWeight.w500)),
                        subtitle: Text(device.modelName ?? '未绑定物模型', style: const TextStyle(color: Color(0xFF757575), fontSize: 12)),
                        trailing: Container(
                          width: 10,
                          height: 10,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: device.isOnline ? const Color(0xFF4CAF50) : const Color(0xFFBDBDBD),
                            boxShadow: device.isOnline
                                ? [BoxShadow(color: const Color(0xFF4CAF50).withOpacity(0.5), blurRadius: 6)]
                                : null,
                          ),
                        ),
                        onTap: () {
                          widget.onSelect(device);
                          Navigator.pop(context);
                        },
                      ),
                    ),
                  );
                },
              ),
            ),
    );
  }
}
