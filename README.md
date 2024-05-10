# Yami - Yet another monkey interpreter
Interpreter for **monkey language** described in [Interpreter in Go](https://interpreterbook.com) by Thorsten Ball. 
As monkey language has no specification some features were added.
## Example

```monkey
// Fibonacci sequence
let n = 10;
let fib = fn (cur, prev, cur_n) {
    if cur_n == n {
        return cur
    }

    return fib(cur + prev, cur, cur_n + 1)
}
  
fib(0, 1, 0)  
```



