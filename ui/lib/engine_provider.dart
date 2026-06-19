import 'dart:convert';
import 'dart:io';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'engine_provider.g.dart';

@riverpod
class EngineLogs extends _$EngineLogs {
  @override
  List<String> build() => [];

  Process? _process;

  void addLog(String log) {
    state = [...state, log];
  }

  Future<void> startEngine() async {
    // Try to find the engine binary.
    // 1. Check environment variable CHRONOS_ENGINE_PATH
    // 2. Check local directory (for dev)
    // 3. Check relative path (for dev)

    String? executable = Platform.environment['CHRONOS_ENGINE_PATH'];

    if (executable == null) {
      final localPath = './chronos-engine';
      final relativePath = '../engine/chronos-engine';

      if (File(localPath).existsSync()) {
        executable = localPath;
      } else if (File(relativePath).existsSync()) {
        executable = relativePath;
      }
    }

    if (executable == null) {
      addLog('Error: Engine binary not found. Set CHRONOS_ENGINE_PATH or place it in the working directory.');
      return;
    }

    try {
      _process = await Process.start(executable, []);
      addLog('Engine process started: $executable');

      _process!.stdout
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen((line) {
        try {
          final json = jsonDecode(line);
          if (json['type'] == 'log') {
            addLog('[${json['level']}] ${json['message']}');
          }
        } catch (e) {
          addLog('Raw: $line');
        }
      });

      _process!.stderr
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen((line) {
        addLog('STDERR: $line');
      });

      _process!.exitCode.then((code) {
        addLog('Engine exited with code $code');
        _process = null;
      });
    } catch (e) {
      addLog('Failed to start engine: $e');
    }
  }

  void sendCommand(String action, [Map<String, dynamic>? params]) {
    if (_process == null) {
      addLog('Error: Engine not running');
      return;
    }
    final cmd = jsonEncode({
      'action': action,
      'params': params ?? {},
    });
    _process!.stdin.writeln(cmd);
  }
}
