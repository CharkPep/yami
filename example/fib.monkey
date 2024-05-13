let fib = fn () {   
    let cache = {}
    let n = 0
    let fib_r = fn (cur, prev, cur_n) {
        cache[cur_n] = cur
        if cur_n == n { 
            return cur
        }

        return fib_r(cur + prev, cur, cur_n + 1)
    }
    return fn(n) {  
        if cache[n] {
            print("Cache hit")
            return cache[n]
        }
   
        return fib_r(1, 0, 0)
    } 
} 

// creating closure under fib
let f = fib()
print(f(6))
print(f(5))
