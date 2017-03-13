# testexpect package

The `testexpect` go package provides some nice assertions to simplify testing.  Failed assertions call `t.Fatalf` to show an error and terminate the test.

## Getting started

Install in the usual fashion using `go get` or your favourite package manager:

    go get -u github.com/stugotech/testexpect 

Create a context per test:

    func TestSomething(t *testing.T) {
        expect := testexpect.NewContext(t)
        answer := CalculateAnswer()
        expect.Equal("answer", answer, 42)
    }

## Methods 

The `Expect` interface has the following methods:

    type Expect interface {
        Nil(name string, actual interface{})
        NotNil(name string, actual interface{})
        NoError(action string, err error)
        DeepEqual(name string, actual interface{}, expected interface{})
        NotDeepEqual(name string, actual interface{}, expected interface{})
        Equal(name string, actual interface{}, expected interface{})
        NotEqual(name string, actual interface{}, expected interface{})
    }