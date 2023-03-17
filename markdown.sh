docker run -it -v `pwd`:/data -v /home/elek:/home/elek --rm pandoc/ubuntu-latex web/content/index.md -o web/template/index.html 
