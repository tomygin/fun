fun fibonaci (n){
    if(n < 2){
      return   1
    }else{
         return (
            fibonaci(n - 1)
            +
            fibonaci(n - 2)
        )
    }
}

print(now())

times := 0

print(fibonaci(2))

print(now())

//TODO:fix fibonaci
