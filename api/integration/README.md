# Integration Testing with Go

This example shows how to do integration testing against a simplistic Go API, using Go tests, and running completely in Docker, so as to avoid dependency issues on the build host or someone's developer machine.

## Test Design

Integration testing is defined as "the testing of various modules of the software under development together as a group. This determines whether or not they function together seamlessly as part of the system or whole." (H/T [CodeAcademy](https://www.codecademy.com/resources/blog/what-is-integration-testing/) for the definition.)

In this exercise, my goal was to create a simple example of an integration test that:

* Was portable, in that it didn't hard-code any links or pointers to things it tested. Thus, it could be used against any environment where the software it tests is deployed.
* Didn't have any host dependencies other than Docker and Make.

To accomplish this, I did the following:

* Wrote the integration test code such that variables like the expected value and the endpoint to be tested are passed in as environment variables (thus avoiding hard-coding anything and following the [12 Factor App](https://12factor.net) model in the process). These values are required by way of using the `require` package from stretchr's [`testify`](https://github.com/stretchr/testify) module.
  ```go
  func TestGetWeb(t *testing.T) {
    URL := os.Getenv("WEB_URL")
    expectedValue := os.Getenv("WEB_EXPECTED_VALUE")
    require.NotEmpty(t, URL)
    require.NotEmpty(t, expectedValue)
  ```
* Bundled the integration test into its own Docker image, such that once built it could be run anywhere that has Docker installed. The Dockerfile simply runs `go test` as its entrypoint.
  ```Dockerfile
  ENTRYPOINT [ "go", "test", "./...", "-v", "-tags", "integration" ]
  ```
  > Note the use of [build constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints), thereby ensuring these test files a) aren't built as part of a normal `go build` run; b) a regular `go test ...` run won't invoke these tests.

## Components

The web API is defined in the [`web/web.go`](web/web.go) file, and the tests are in the [`integration/web_test.go`](integration/web_test.go) file.

There are two Dockerfiles: [`Dockerfile.web`](Dockerfile.web), which builds the image for the web API, and [`Dockerfile.int-test`](Dockerfile.int-test), which packages the integration tests.

## Executing Test

To execute the test, all you need to do is run `make run-integration-test` (or if you're on a Mac and have installed Make via Homebrew, `gmake run-integration-test`). This:

* Builds Docker images for both the web and test containers.
* Creates a named Docker network.
* Starts the web API container in the network.
* Starts the test container in the network, while passing the name of the API container as the environment variable used for determining the test endpoint. The image will automatically exit when the test completes.
* Upon successful or failed test completion, stops the web API container, and deletes the network.

>**NOTE:** That last point required some interesting trickery in the Makefile, namely the use of the [`.ONESHELL`](https://www.gnu.org/software/make/manual/html_node/One-Shell.html) directive and Bash's [trap](https://ss64.com/bash/trap.html) command. This essentially mimics a global teardown method that is typical in testing frameworks. However, it requires Make version 3.8.2 or higher, so for you Mac folks, make sure you install that since OSX ships with version 3.8.1 by default. I recommend the [Homebrew formula](https://formulae.brew.sh/formula/make), which is included in the Brewfile at the root of this repository.
