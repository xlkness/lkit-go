package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Tree(dir string) {
	tree(dir, 1, true)
}

func tree(dir string, deepth int, prelast bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("查看目录%v所有文件错误:%v\n", dir, err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", filepath.Base(dir))
	for i, v := range files {
		if i == len(files)-1 {
			treeGraph(deepth, prelast, true)
			prelast = true
		} else {
			treeGraph(deepth, prelast, false)
			prelast = false
		}
		if v.IsDir() {
			tree(dir+"/"+v.Name(), deepth+1, prelast)
		} else {
			fmt.Printf("%v\n", filepath.Base(v.Name()))
		}
	}
}

func treeGraph(deepth int, prelast, last bool) {
	if deepth == 1 {
		if last {
			fmt.Printf("└── ")
		} else {
			fmt.Printf("│── ")
		}
	} else {
		if prelast && last {
			for i := 0; i < deepth-1; i++ {
				fmt.Printf("    ")
			}
			fmt.Printf("└── ")
		} else if !prelast && last {
			fmt.Printf("│   ")
			for i := 0; i < deepth-2; i++ {
				fmt.Printf(" ")
			}
			fmt.Printf("└── ")
		} else if !prelast && !last {
			fmt.Printf("│   ")
			for i := 0; i < deepth-2; i++ {
				fmt.Printf(" ")
			}
			fmt.Printf("│── ")
		} else if prelast && !last {
			for i := 0; i < deepth-1; i++ {
				fmt.Printf("    ")
			}
			fmt.Printf("│── ")
		}
	}
}
