import 'package:flutter/material.dart';

class QuickActions extends StatelessWidget {
  final VoidCallback onViewData;
  final VoidCallback onSetProperty;
  final VoidCallback onInvokeService;

  const QuickActions({
    super.key,
    required this.onViewData,
    required this.onSetProperty,
    required this.onInvokeService,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
        boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 10, offset: Offset(0, -2))],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _ActionTile(icon: Icons.data_usage, label: '查看设备数据', onTap: onViewData),
          _ActionTile(icon: Icons.settings, label: '下发属性', onTap: onSetProperty),
          _ActionTile(icon: Icons.play_arrow, label: '调用服务', onTap: onInvokeService),
        ],
      ),
    );
  }
}

class _ActionTile extends StatefulWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  const _ActionTile({required this.icon, required this.label, required this.onTap});

  @override
  State<_ActionTile> createState() => _ActionTileState();
}

class _ActionTileState extends State<_ActionTile> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _pressed = true),
      onTapUp: (_) {
        setState(() => _pressed = false);
        widget.onTap();
      },
      onTapCancel: () => setState(() => _pressed = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        margin: const EdgeInsets.only(bottom: 8),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          color: _pressed ? const Color(0xFFE3F2FD) : const Color(0xFFF5F5F5),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(widget.icon, color: const Color(0xFF2196F3)),
            const SizedBox(width: 12),
            Text(widget.label, style: const TextStyle(color: Color(0xFF212121), fontSize: 15)),
          ],
        ),
      ),
    );
  }
}
