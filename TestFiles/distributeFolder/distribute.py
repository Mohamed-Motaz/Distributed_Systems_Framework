"""
This file is responsible for receiving the jobContent, reading it, and writing the distribute
"""


#read the distribute.txt file that the system should create before running me.
#distribute.txt should be in my same directory if I am a file -- ./distribute.txt
#it should be up a level if I am a folder ././distribute.txt

f = open("././distribute.txt", "r+")
#assume distribute.txt contains no job content

print("-----"  + f.read() + "----")

#empty distribute.txt file
f.seek(0)
f.truncate()


#now write the tasks as a list
tasks = ['optionalFile1.txt']
toWrite = ",".join(tasks)

f.write(toWrite)

f.close()