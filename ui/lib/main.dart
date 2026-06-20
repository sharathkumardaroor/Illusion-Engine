import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:file_picker/file_picker.dart';
import 'package:intl/intl.dart';
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

class ChronosWorkspace extends ConsumerStatefulWidget {
  const ChronosWorkspace({super.key});

  @override
  ConsumerState<ChronosWorkspace> createState() => _ChronosWorkspaceState();
}

class _ChronosWorkspaceState extends ConsumerState<ChronosWorkspace> {
  String? sourceDir;
  String? outputDir;
  String engineMode = 'Deterministic';
  String cadence = 'Medium';
  DateTime startDate = DateTime.now().subtract(const Duration(days: 90));
  DateTime endDate = DateTime.now();

  final TextEditingController baseUrlController = TextEditingController();
  final TextEditingController apiKeyController = TextEditingController();
  final TextEditingController modelController = TextEditingController();
  final TextEditingController focusAreaController = TextEditingController();
  final TextEditingController struggleAreaController = TextEditingController();

  bool humanErrors = true;
  bool astPhasing = true;
  bool depAlignment = true;
  bool branches = true;

  @override
  Widget build(BuildContext context) {
    final engineState = ref.watch(engineStateProvider);
    final logs = engineState['logs'] as List<String>;
    final isRunning = engineState['status'] == 'running' || engineState['status'] == 'preparing-test';

    return Scaffold(
      body: Row(
        children: [
          // Left Pane: Configuration
          Expanded(
            flex: 3,
            child: Container(
              color: Theme.of(context).colorScheme.surfaceVariant.withOpacity(0.3),
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Wrap(
                      alignment: WrapAlignment.spaceBetween,
                      crossAxisAlignment: WrapCrossAlignment.center,
                      spacing: 16,
                      runSpacing: 16,
                      children: [
                        Text('Chronos Workspace', style: Theme.of(context).textTheme.headlineMedium),
                        OutlinedButton.icon(
                          onPressed: isRunning ? null : () => ref.read(engineStateProvider.notifier).runTestSimulation(),
                          icon: const Icon(Icons.science),
                          label: const Text('Run Test Simulation'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 32),

                    _sectionTitle('Source Project Directory'),
                    const SizedBox(height: 8),
                    _directoryPicker(
                      path: sourceDir,
                      onPick: () async {
                        String? result = await FilePicker.platform.getDirectoryPath();
                        if (result != null) {
                          setState(() {
                            sourceDir = result;
                            outputDir ??= '${result}_chronos';
                          });
                          ref.read(engineStateProvider.notifier).scan(result);
                        }
                      },
                      placeholder: 'Select source project folder',
                    ),

                    if (engineState['scanResult'] != null) ...[
                      const SizedBox(height: 12),
                      _scanResultCard(engineState['scanResult']),
                    ],

                    const SizedBox(height: 24),
                    _sectionTitle('Output Directory'),
                    const SizedBox(height: 8),
                    _directoryPicker(
                      path: outputDir,
                      onPick: () async {
                        String? result = await FilePicker.platform.getDirectoryPath();
                        if (result != null) setState(() => outputDir = result);
                      },
                      placeholder: 'Select where to create revamp',
                    ),

                    const SizedBox(height: 24),
                    _sectionTitle('Engine Config'),
                    DropdownButtonFormField<String>(
                      value: engineMode,
                      items: ['Deterministic', 'Free Cloud AI', 'Local AI (Ollama)', 'Custom API']
                          .map((e) => DropdownMenuItem(value: e, child: Text(e)))
                          .toList(),
                      onChanged: (val) {
                        setState(() {
                          engineMode = val!;
                          if (val == 'Free Cloud AI') {
                            baseUrlController.text = 'https://text.pollinations.ai/';
                            modelController.text = 'gpt-4o';
                          } else if (val == 'Local AI (Ollama)') {
                            baseUrlController.text = 'http://localhost:11434/v1';
                            modelController.text = 'llama3';
                          }
                        });
                      },
                      decoration: const InputDecoration(border: OutlineInputBorder()),
                    ),
                    if (engineMode == 'Custom API' || engineMode == 'Free Cloud AI' || engineMode == 'Local AI (Ollama)') ...[
                      const SizedBox(height: 12),
                      TextField(controller: baseUrlController, decoration: const InputDecoration(labelText: 'Base URL', border: OutlineInputBorder())),
                      if (engineMode == 'Custom API') ...[
                        const SizedBox(height: 12),
                        TextField(controller: apiKeyController, decoration: const InputDecoration(labelText: 'API Key', border: OutlineInputBorder())),
                      ],
                      const SizedBox(height: 12),
                      TextField(controller: modelController, decoration: const InputDecoration(labelText: 'Model Name', border: OutlineInputBorder())),
                    ],

                    const SizedBox(height: 24),
                    _sectionTitle('Timeline Configuration'),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<String>(
                      value: cadence,
                      items: ['Low', 'Medium', 'High']
                          .map((e) => DropdownMenuItem(value: e, child: Text('Commit Cadence: $e')))
                          .toList(),
                      onChanged: (val) => setState(() => cadence = val!),
                      decoration: const InputDecoration(border: OutlineInputBorder()),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: () async {
                              final picked = await showDatePicker(context: context, initialDate: startDate, firstDate: DateTime(2000), lastDate: DateTime(2100));
                              if (picked != null) setState(() => startDate = picked);
                            },
                            icon: const Icon(Icons.calendar_today),
                            label: Text('Start: ${DateFormat('yyyy-MM-dd').format(startDate)}'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: OutlinedButton.icon(
                            onPressed: () async {
                              final picked = await showDatePicker(context: context, initialDate: endDate, firstDate: DateTime(2000), lastDate: DateTime(2100));
                              if (picked != null) setState(() => endDate = picked);
                            },
                            icon: const Icon(Icons.calendar_today),
                            label: Text('End: ${DateFormat('yyyy-MM-dd').format(endDate)}'),
                          ),
                        ),
                      ],
                    ),

                    const SizedBox(height: 24),
                    _sectionTitle('Narrative Focus'),
                    TextField(controller: focusAreaController, decoration: const InputDecoration(labelText: 'Focus Area (e.g. backend refactor)')),
                    TextField(controller: struggleAreaController, decoration: const InputDecoration(labelText: 'Struggle Area (e.g. race conditions)')),

                    const SizedBox(height: 24),
                    _sectionTitle('Realism Injectors'),
                    CheckboxListTile(title: const Text('Human Error Injection'), value: humanErrors, onChanged: (val) => setState(() => humanErrors = val!)),
                    CheckboxListTile(title: const Text('AST Code Phasing'), value: astPhasing, onChanged: (val) => setState(() => astPhasing = val!)),
                    CheckboxListTile(title: const Text('Dependency Alignment'), value: depAlignment, onChanged: (val) => setState(() => depAlignment = val!)),
                    CheckboxListTile(title: const Text('Branch Simulation'), value: branches, onChanged: (val) => setState(() => branches = val!)),

                    const SizedBox(height: 24),
                    if (engineState['estimate'] != null) ...[
                      _estimateCard(engineState['estimate']),
                      const SizedBox(height: 24),
                    ],

                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      height: 50,
                      child: ElevatedButton(
                        style: ElevatedButton.styleFrom(backgroundColor: Colors.blueAccent, foregroundColor: Colors.white),
                        onPressed: sourceDir == null || outputDir == null || isRunning ? null : () {
                          ref.read(engineStateProvider.notifier).execute({
                            'sourceDir': sourceDir,
                            'outputDir': outputDir,
                            'useAI': engineMode != 'Deterministic',
                            'provider': engineMode,
                            'baseUrl': baseUrlController.text,
                            'apiKey': apiKeyController.text,
                            'model': modelController.text,
                            'focusArea': focusAreaController.text,
                            'struggleArea': struggleAreaController.text,
                            'startDate': startDate.toIso8601String(),
                            'endDate': endDate.toIso8601String(),
                            'cadence': cadence,
                            'humanErrors': humanErrors,
                            'astPhasing': astPhasing,
                            'depAlignment': depAlignment,
                            'branches': branches,
                          });
                        },
                        child: Text(isRunning ? 'Executing Revamp...' : 'Execute Chronos'),
                      ),
                    ),
                    const SizedBox(height: 12),
                    if (sourceDir != null && !isRunning)
                      Center(
                        child: TextButton(
                          onPressed: () {
                             ref.read(engineStateProvider.notifier).estimate({
                                'sourceDir': sourceDir,
                                'useAI': engineMode != 'Deterministic',
                                'branches': branches,
                                'cadence': cadence,
                             });
                          },
                          child: const Text('Refresh Estimate'),
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ),
          const VerticalDivider(width: 1),
          // Right Pane: Live State & Logs
          Expanded(
            flex: 2,
            child: Container(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('State Summary', style: Theme.of(context).textTheme.headlineSmall),
                  const SizedBox(height: 24),
                  _stateCard('Source', sourceDir == null ? '-' : '${engineState['commitsBefore']} commits'),
                  const SizedBox(height: 12),
                  _stateCard('Output', '${engineState['commitsAfter']} commits'),
                  const SizedBox(height: 12),
                  _stateCard('Status', engineState['status'].toString().toUpperCase()),
                  if (engineState['verified'] == true) ...[
                    const SizedBox(height: 12),
                    const Row(
                      children: [
                        Icon(Icons.check_circle, color: Colors.green),
                        SizedBox(width: 8),
                        Text('Verification Clean', style: TextStyle(color: Colors.green, fontWeight: FontWeight.bold)),
                      ],
                    ),
                  ],
                  const SizedBox(height: 24),
                  const Text('Live Engine Log', style: TextStyle(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  Expanded(
                    child: Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.black,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: ListView.builder(
                        itemCount: logs.length,
                        itemBuilder: (context, index) => Text(
                          logs[index],
                          style: const TextStyle(color: Colors.greenAccent, fontFamily: 'JetBrains Mono', fontSize: 11),
                        ),
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

  Widget _sectionTitle(String title) => Text(title, style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.blueAccent));

  Widget _directoryPicker({required String? path, required VoidCallback onPick, required String placeholder}) => Row(
    children: [
      Expanded(
        child: Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(color: Colors.black26, borderRadius: BorderRadius.circular(8)),
          child: Text(path ?? placeholder, overflow: TextOverflow.ellipsis),
        ),
      ),
      const SizedBox(width: 12),
      ElevatedButton(onPressed: onPick, child: const Text('Pick Folder')),
    ],
  );

  Widget _stateCard(String label, String value) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
    decoration: BoxDecoration(color: Colors.white10, borderRadius: BorderRadius.circular(8)),
    child: Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label),
        const SizedBox(width: 8),
        Expanded(child: Text(value, textAlign: TextAlign.right, style: const TextStyle(fontWeight: FontWeight.bold, fontFamily: 'JetBrains Mono'))),
      ],
    ),
  );

  Widget _scanResultCard(Map<String, dynamic> scan) => Container(
    padding: const EdgeInsets.all(16),
    decoration: BoxDecoration(color: Colors.blueGrey.withOpacity(0.1), borderRadius: BorderRadius.circular(8), border: Border.all(color: Colors.blueAccent.withOpacity(0.3))),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Icon(Icons.info_outline, size: 16, color: Colors.blueAccent),
            const SizedBox(width: 8),
            Text('Project Scan', style: TextStyle(fontWeight: FontWeight.bold, color: Colors.blueAccent.shade100)),
          ],
        ),
        const SizedBox(height: 12),
        _scanRow('Files/Folders', '${scan['fileCount']} / ${scan['folderCount']}'),
        _scanRow('Total Size', '${(scan['sizeMb'] as double).toStringAsFixed(2)} MB'),
        if (scan['isGit'] == true) ...[
          const Divider(height: 20),
          _scanRow('Git Commits', '${scan['commitCount']}'),
          _scanRow('Branches', '${scan['branchCount']}'),
          _scanRow('First Commit', '${scan['firstCommit']}'),
          _scanRow('Latest Commit', '${scan['latestCommit']}'),
        ],
      ],
    ),
  );

  Widget _scanRow(String label, String value) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 2),
    child: Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: const TextStyle(fontSize: 12, color: Colors.white70)),
        Text(value, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, fontFamily: 'JetBrains Mono')),
      ],
    ),
  );

  Widget _estimateCard(Map<String, dynamic> est) => Container(
    padding: const EdgeInsets.all(16),
    decoration: BoxDecoration(color: Colors.orange.withOpacity(0.05), borderRadius: BorderRadius.circular(8), border: Border.all(color: Colors.orange.withOpacity(0.3))),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Icon(Icons.analytics_outlined, size: 16, color: Colors.orange),
            const SizedBox(width: 8),
            Text('Revamp Estimate', style: TextStyle(fontWeight: FontWeight.bold, color: Colors.orange.shade100)),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _estItem(Icons.history, '${est['commits']}', 'Commits'),
            _estItem(Icons.account_tree, '${est['branches']}', 'Branches'),
            _estItem(Icons.timer, '${est['runtime']}', 'Time'),
          ],
        ),
      ],
    ),
  );

  Widget _estItem(IconData icon, String val, String label) => Column(
    children: [
      Icon(icon, size: 20, color: Colors.white54),
      const SizedBox(height: 4),
      Text(val, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
      Text(label, style: const TextStyle(fontSize: 10, color: Colors.white54)),
    ],
  );
}
