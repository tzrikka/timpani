# Autobahn Testsuite

The [**Autobahn**|Testsuite](https://github.com/crossbario/autobahn-testsuite) provides a fully automated test suite to verify client and server implementations of [The WebSocket Protocol](https://datatracker.ietf.org/doc/html/rfc6455) for specification conformance and implementation robustness.

Instructions:

1. Start Autobahn's Docker container:

   ```shell
   docker run -it --rm \
       -v "${PWD}/config:/config" \
       -v "${PWD}/reports:/reports" \
       -p 9001:9001 \
       --name fuzzingserver \
       crossbario/autobahn-testsuite
   ```

2. Run Timpani's WebSocket client tester:

   ```shell
   go run ./wstest
   ```

3. Review the results report in: `reports/clients/index.html`
