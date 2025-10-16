fun Person(name){
    fun Say(what){
        print(name," ","say"," ",what)
    }
    return this
}

p := Person('Gin')
p.Say('Good!')

print(VERSION)
