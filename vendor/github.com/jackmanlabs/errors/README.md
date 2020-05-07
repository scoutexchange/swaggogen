# errors
This package has become my default error package for the purpose of passing
errors up the stack. It allows the programmer to define a human-readable
explanation of the error and a simple stack trace. If I were to include anything
more, it would be a list of variables and their values in the scope in the top
most stack frame being tracked. Security is potentially an issue, so I haven't
investigated that possibility at any length.

## 30 March 2018
I've revised this package to store the root error and the individual stack data
separately. This was to enable (eventually) a Root() method that would allow us
to check the root error for specific handling up the stack.

The result is (IMHO) cleaner code. The only breaking change was to make Stack()
return nil on nil argument. This pattern is something I've seen preferred by my
coworkers, and it doesn't affect my own personal use of this package. 
