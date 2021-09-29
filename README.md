# xkcd_searchtool
a search engine for XKCD

## What is it?

- This tool can **crawling the comic** **transcripts** from [XKCD.com](http://xkcd.com)
- Users can search a comic using key words written in a transcript and Get the most 5 relevant comic urls so they can use it to get to the comic they want.

## How to use it?

You should use this tool in your command line.

There will be two kinds of args you need to care about.

- -u true (or false) : default is true, it's used for control the tool of updating index or not.
- -kw (string): you need to type in your keywords for searching.(separited with space, need to be surround with "" )

## Example in my PC
You can download the exe in Release Page.



```shell
PS C:\Users\wgg96\Desktop> .\xkcd_searchtool.exe -kw="sheep"
start to update index
update finished
start to search
Search Result:
1: https://xkcd.com/571/info.0.json
print finished bye~
```
You can go to web browser to search the url and get what you want ~~ 

Have fun with it~
