let end = 10;
let loop = fn (cur, incr) {
    if (cur == end) {
        return
    }

    print("Iteration: " + cur);
    return loop(cur + incr, incr);
};

loop(0, 1)