package main

func main() {
	app := App{}
	env := getEnv()
	app.Initialize(env)
	app.Run(":8080")
}
