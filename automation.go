package main


import(

    "os/exec"
    "runtime"
    "fmt"
    "log"
    
)

func main(){    
	
	fmt.Println("Installing automation ....")
	
	switch os := runtime.GOOS; os {
	case "windows":
		exec.Command("cmd", "/C", "install.bat")
    case "linux":
    	{ 
		 out, err := exec.Command("/bin/bash","install").Output()
		 if err != nil {
		 log.Fatal(err)
		 }
		 fmt.Println(string(out))
		}		
	default:
     { 
			out, err := exec.Command("/bin/bash","install").Output()
			if err != nil {
			log.Fatal(err)
			}
			fmt.Println(string(out))
	 }	
	}
	
	fmt.Println("Installing ended ....") 	       
}

