let count = fn() { 
    let counter = 0; 
    return fn() { 
        counter = counter + 1; 
        return counter 
    }
}

let c = count()

print(c())
print(c())
print(c())
print(c())
print(c())

