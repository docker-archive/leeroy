stage "test"
def testStep = golangTester(
  package: "github.com/docker/leeroy",
  go_version: "1.6.2",
  max_warnings: 1, // run gofmt -w -s if new warnings come up
).call()
