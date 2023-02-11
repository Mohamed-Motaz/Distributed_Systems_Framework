"""
This file is responsible for receiving an optionalFilePath, reading it, and writing it back out twice
"""

import os

#read the process.txt file that the system should create before running me.
#process.txt should be in my same directory if I am a file -- ./process.txt
#it should be up a level if I am a folder ./../process.txt

f = open("process.txt", "r+")

optionalFilePath = f.read()
print("optionalFilePath: " + optionalFilePath)

optionalFile = open(optionalFilePath, "r")
optionalFileData = optionalFile.readlines()
optionalFile.close()


#empty process.txt file
f.seek(0)
f.truncate()
#now write the data back to process.txt
for data in optionalFileData:
    f.write(data)
    f.write(data)

f.close()





