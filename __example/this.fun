fun Person(name){
    that = this
    fun Say(what){
        print(that.name,"say",what)
    }
    return this
}

p = Person('Gin')
p.Say('Good!')
