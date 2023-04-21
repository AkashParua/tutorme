import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

void main() {
  runApp(const MaterialApp(home: MainApp()));
}

class MainApp extends StatefulWidget {
  const MainApp({super.key});

  @override
  MainAppState createState() => MainAppState();
}

class MainAppState extends State<MainApp> {
  final TextEditingController _controller = TextEditingController();
  String _responseText = '';
  Future<void> _sendRequest(String endpoint, String question) async {
    final url = Uri.http('127.0.0.1:5000', endpoint, {'question': question});
    // print(url.toString());
    final response = await http.get(url);

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      //print(data);
      // print(jsonResponse);
      //Map<String, dynamic> jsonData = jsonResponse;
      //print(endpoint);
      setState(() {
        if (endpoint == '/search') {
          _responseText = data;
        } else {
          _responseText = data['response'];
        }
        //print(_responseText);
      });
    } else {
      setState(() {
        _responseText = 'Request failed with status: ${response.statusCode}.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('TutorGPT'),
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TextField(
              controller: _controller,
              decoration: const InputDecoration(
                hintText: 'Question',
              ),
            ),
            const SizedBox(height: 16.0),
            ElevatedButton(
                onPressed: () async {
                  setState(() {
                    _responseText = "";
                  });
                  final input = _controller.text;
                  await _sendRequest('/search', input);
                },
                child: const Text('Search Book')),
            ElevatedButton(
                onPressed: () async {
                  setState(() {
                    _responseText = "....processing ....";
                  });

                  final input = _controller.text;
                  await _sendRequest('search_and_ask', input);
                },
                child: const Text('ask GPT from book')),
            const SizedBox(height: 16.0),
            Text(_responseText),
          ],
        ),
      ),
    );
  }
}
