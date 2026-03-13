import 'package:flutter/material.dart';

class DeviceDataCard extends StatelessWidget {
  final Map<String, dynamic> data;

  const DeviceDataCard({super.key, required this.data});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.06), blurRadius: 8, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('设备数据', style: TextStyle(color: Color(0xFF2196F3), fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          ...data.entries.map((e) => Padding(
                padding: const EdgeInsets.symmetric(vertical: 4),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(e.key, style: const TextStyle(color: Color(0xFF757575))),
                    Text('${e.value}', style: const TextStyle(color: Color(0xFF212121), fontWeight: FontWeight.w500)),
                  ],
                ),
              )),
        ],
      ),
    );
  }
}
