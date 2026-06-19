import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'engine_provider.dart';

void main() {
  runApp(const ProviderScope(child: MyApp()));
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Chronos',
      theme: ThemeData(
        brightness: Brightness.dark,
        useMaterial3: true,
        fontFamily: 'Inter',
        colorSchemeSeed: Colors.blue,
      ),
      home: const ChronosWorkspace(),
    );
  }
}

class ChronosWorkspace extends ConsumerWidget {
  const ChronosWorkspace({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final logs = ref.watch(engineLogsProvider);

    return Scaffold(
      body: Row(
        children: [
          // Left Pane: Configuration
          Expanded(
            flex: 1,
            child: Container(
              color: Theme.of(context).colorScheme.surfaceVariant,
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Configuration',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 20),
                  const Text('Engine Settings'),
                  const Spacer(),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      ElevatedButton(
                        onPressed: () {
                          ref.read(engineLogsProvider.notifier).startEngine();
                        },
                        child: const Text('Start Engine'),
                      ),
                      ElevatedButton(
                        onPressed: () {
                          ref.read(engineLogsProvider.notifier).sendCommand('ping');
                        },
                        child: const Text('Ping'),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const VerticalDivider(width: 1),
          // Right Pane: Logs & Status
          Expanded(
            flex: 2,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Summary & Logs',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 20),
                  Expanded(
                    child: Container(
                      color: Colors.black,
                      width: double.infinity,
                      padding: const EdgeInsets.all(8),
                      child: ListView.builder(
                        itemCount: logs.length,
                        itemBuilder: (context, index) {
                          return Text(
                            logs[index],
                            style: const TextStyle(
                              color: Colors.greenAccent,
                              fontFamily: 'JetBrains Mono',
                              fontSize: 12,
                            ),
                          );
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
