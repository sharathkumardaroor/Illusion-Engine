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

  Future<void> execute(Map<String, dynamic> config) async {
    // Kill existing process and cancel subscriptions if running
    if (_process != null) {
      _process!.kill();
      _process = null;
    }
    await _stdoutSub?.cancel();
    await _stderrSub?.cancel();
    _stdoutSub = null;
    _stderrSub = null;

    state = {
      ...state,
      'logs': <String>[],
      'status': 'running',
    };

    if (kIsWeb) {
      addLog('Error: Engine execution not supported on Web.');
      state = {...state, 'status': 'error'};
      return;
    }

    String? executable = Platform.environment['CHRONOS_ENGINE_PATH'];
    if (executable == null) {
      final localPath = './chronos-engine';
      final relativePath = '../engine/chronos-engine';
      if (File(localPath).existsSync()) {
        executable = localPath;
      } else if (File(relativePath).existsSync()) {
        executable = relativePath;
      } else if (File('$localPath.exe').existsSync()) {
        executable = '$localPath.exe';
      } else if (File('$relativePath.exe').existsSync()) {
        executable = '$relativePath.exe';
      }
    }

    if (executable == null) {
      addLog('Error: Engine binary not found.');
      state = {...state, 'status': 'error'};
      return;
    }

    try {
      _process = await Process.start(executable, []);
      addLog('Engine started: $executable');

      _stdoutSub = _process!.stdout
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen((line) {
        try {
          final json = jsonDecode(line);
          if (json['type'] == 'log') {
            addLog('[${json['level']}] ${json['message']}');
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
          } else if (json['type'] == 'estimate') {
             final payload = json['payload'];
             addLog('Estimate: ${payload['commits']} commits, ${payload['runtime']} runtime');
          }
        } catch (e) {
          addLog('Raw: $line');
        }
      });

      _stderrSub = _process!.stderr
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen((line) {
        addLog('STDERR: $line');
      });

      _process!.stdin.writeln(jsonEncode({
        'action': 'execute',
        'params': config,
      }));
      // Close stdin to signal end of commands and allow engine to exit after task
      await _process!.stdin.close();

      _process!.exitCode.then((code) {
        addLog('Engine exited with code $code');
        _process = null;
      });
    } catch (e) {
      addLog('Failed to start engine: $e');
      state = {...state, 'status': 'error'};
    }
  }
}
