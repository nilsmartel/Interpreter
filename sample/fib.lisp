(fun fib (n)
     (if (== n 0) 0
       (if (== n 1) 1
         (fib (- n 1) (- n 2)))))

(print (fib 0))
(print (fib 1))
(print (fib 2))
(print (fib 3))
(print (fib 4))
(print (fib 5))
(print (fib 6))
(print (fib 8))
(print (fib 9))
