import 'package:flutter_test/flutter_test.dart';
import 'package:ui/main.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

void main() {
  testWidgets('Chronos Workspace basic smoke test', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    await tester.pumpWidget(const ProviderScope(child: MyApp()));

    // Verify that our workspace is displayed.
    expect(find.text('Chronos Workspace'), findsOneWidget);
    expect(find.text('State Summary'), findsOneWidget);
    expect(find.text('Execute Chronos'), findsOneWidget);
  });
}
