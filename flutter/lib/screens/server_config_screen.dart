import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../config.dart';
import '../widgets/animated_button.dart';

class ServerConfigScreen extends StatefulWidget {
  const ServerConfigScreen({super.key});

  @override
  State<ServerConfigScreen> createState() => _ServerConfigScreenState();
}

class _ServerConfigScreenState extends State<ServerConfigScreen> with SingleTickerProviderStateMixin {
  final _ipController = TextEditingController();
  final _portController = TextEditingController(text: '${AppConfig.defaultPort}');
  bool _loading = false;
  String? _error;
  late AnimationController _animController;
  late Animation<double> _fadeAnim;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(vsync: this, duration: const Duration(milliseconds: 600));
    _fadeAnim = CurvedAnimation(parent: _animController, curve: Curves.easeOut);
    _animController.forward();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  Future<void> _testConnection() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    final ip = _ipController.text.trim();
    final port = _portController.text.trim();
    if (ip.isEmpty) {
      setState(() {
        _error = '请输入服务器地址';
        _loading = false;
      });
      return;
    }

    final url = 'http://$ip:$port';
    await ApiService.setServerUrl(url);

    try {
      await ApiService.request('GET', '/health');
      if (mounted) Navigator.pushReplacementNamed(context, '/login');
    } catch (e) {
      setState(() {
        _error = '连接失败: $e';
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: Center(
        child: FadeTransition(
          opacity: _fadeAnim,
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.cloud, size: 80, color: Color(0xFF2196F3)),
                const SizedBox(height: 24),
                const Text('配置服务器', style: TextStyle(fontSize: 28, fontWeight: FontWeight.bold, color: Color(0xFF212121))),
                const SizedBox(height: 48),
                TextField(
                  controller: _ipController,
                  style: const TextStyle(color: Color(0xFF212121)),
                  decoration: InputDecoration(
                    labelText: 'IP 地址',
                    hintText: '192.168.1.100',
                    prefixIcon: const Icon(Icons.dns_outlined, color: Color(0xFF2196F3)),
                    filled: true,
                    fillColor: const Color(0xFFF5F5F5),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
                    focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: Color(0xFF2196F3), width: 2)),
                  ),
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _portController,
                  style: const TextStyle(color: Color(0xFF212121)),
                  decoration: InputDecoration(
                    labelText: '端口',
                    prefixIcon: const Icon(Icons.numbers, color: Color(0xFF2196F3)),
                    filled: true,
                    fillColor: const Color(0xFFF5F5F5),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
                    focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: Color(0xFF2196F3), width: 2)),
                  ),
                ),
                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(_error!, style: const TextStyle(color: Color(0xFFE53935))),
                ],
                const SizedBox(height: 24),
                SizedBox(
                  width: double.infinity,
                  child: AnimatedButton(
                    onPressed: _loading ? null : _testConnection,
                    child: _loading
                        ? const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                        : const Text('连接', style: TextStyle(fontSize: 16, color: Colors.white, fontWeight: FontWeight.w600)),
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
