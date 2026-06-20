import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'engine_provider.g.dart';

@riverpod
class EngineState extends _$EngineState {
  @override
  Map<String, dynamic> build() => {
    'logs': <String>[],
    'status': 'idle',
    'verified': false,
    'commitsBefore': 0,
    'commitsAfter': 0,
    'outputPath': '',
    'reportPath': '',
    'testSourcePath': null,
    'scanResult': null,
    'estimate': null,
  };

  Process? _process;
  StreamSubscription? _stdoutSub;
  StreamSubscription? _stderrSub;

  void addLog(String log) {
    state = {
      ...state,
      'logs': [...(state['logs'] as List<String>), log],
    };
  }

  Future<void> _ensureProcess() async {
    if (_process != null) return;

    String? executable = Platform.environment['CHRONOS_ENGINE_PATH'];
    if (executable == null) {
      final localPath = './chronos-engine';
      final relativePath = '../engine/chronos-engine';
      if (File(localPath).existsSync()) executable = localPath;
      else if (File(relativePath).existsSync()) executable = relativePath;
      else if (File('$localPath.exe').existsSync()) executable = '$localPath.exe';
      else if (File('$relativePath.exe').existsSync()) executable = '$relativePath.exe';
    }

    if (executable == null) throw Exception('Engine binary not found');

    _process = await Process.start(executable, []);
    _stdoutSub = _process!.stdout.transform(utf8.decoder).transform(const LineSplitter()).listen((line) {
      try {
        final json = jsonDecode(line);
        if (json['type'] == 'log') {
          addLog('[${json['level']}] ${json['message']}');
          if (json['payload'] != null && json['payload']['path'] != null) {
            state = {...state, 'testSourcePath': json['payload']['path']};
          }
        } else if (json['type'] == 'state') {
          final payload = json['payload'];
          state = {
            ...state,
            'status': payload['status'],
            'verified': payload['verified'] ?? false,
            'commitsBefore': payload['before'] ?? 0,
            'commitsAfter': payload['after'] ?? 0,
            'outputPath': payload['output_path'] ?? '',
            'reportPath': payload['report_path'] ?? '',
          };
        } else if (json['type'] == 'scan_result') {
          state = {...state, 'scanResult': json['payload']};
        } else if (json['type'] == 'estimate') {
          state = {...state, 'estimate': json['payload']};
        }
      } catch (e) {
        addLog('Raw: $line');
      }
    });

    _stderrSub = _process!.stderr.transform(utf8.decoder).transform(const LineSplitter()).listen((line) => addLog('STDERR: $line'));

    _process!.exitCode.then((code) {
      addLog('Engine exited with code $code');
      _process = null;
      if (state['status'] == 'running') state = {...state, 'status': 'error'};
    });
  }

  Future<void> runTestSimulation() async {
    state = {...state, 'status': 'preparing-test', 'logs': <String>[]};
    try {
      await _ensureProcess();
      _process!.stdin.writeln(jsonEncode({'action': 'test-prep', 'params': {}}));

      // Wait for testSourcePath to be populated
      while (state['testSourcePath'] == null) {
        await Future.delayed(const Duration(milliseconds: 100));
      }

      final source = state['testSourcePath'];
      await scan(source);

      final output = '${source}_revamped';

      addLog('Executing simulation: $source -> $output');
      execute({
        'sourceDir': source,
        'outputDir': output,
        'useAI': false,
        'engineMode': 'Deterministic',
      });
    } catch (e) {
      addLog('Test Prep Failed: $e');
      state = {...state, 'status': 'error'};
    }
  }

  Future<void> scan(String path) async {
    try {
      await _ensureProcess();
      _process!.stdin.writeln(jsonEncode({
        'action': 'scan',
        'params': {'path': path},
      }));
    } catch (e) {
      addLog('Scan failed: $e');
    }
  }

  Future<void> estimate(Map<String, dynamic> config) async {
    try {
      await _ensureProcess();
      _process!.stdin.writeln(jsonEncode({
        'action': 'estimate',
        'params': config,
      }));
    } catch (e) {
      addLog('Estimate failed: $e');
    }
  }

  Future<void> execute(Map<String, dynamic> config) async {
    if (_process != null && state['status'] != 'preparing-test') {
      // Keep process alive if we are just switching from idle to running
      // But if it's already running a different execution, we might need to restart it?
      // Actually, the engine is designed to handle multiple commands now.
      // However, execute usually implies a fresh start for the engine's internal state if it had any.
      // But my engine main.go loop handles multiple commands.
    }

    state = {
      ...state,
      'logs': state['status'] == 'preparing-test' ? state['logs'] : <String>[],
      'status': 'running',
      'verified': false,
      'commitsBefore': 0,
      'commitsAfter': 0,
      'outputPath': '',
      'reportPath': '',
    };

    if (kIsWeb) {
      addLog('Error: Engine execution not supported on Web.');
      state = {...state, 'status': 'error'};
      return;
    }

    try {
      await _ensureProcess();
      _process!.stdin.writeln(jsonEncode({
        'action': 'execute',
        'params': config,
      }));
      await _process!.stdin.close();
    } catch (e) {
      addLog('Failed to start engine: $e');
      state = {...state, 'status': 'error'};
    }
  }
}
