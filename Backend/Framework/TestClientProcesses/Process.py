import json

filePath = "./process.txt"

def process():
    f = open(filePath, 'r+')

    taskContentFilePath = f.read().replace("\n", "").strip()


    taskContentF = open(taskContentFilePath, "r")
    file_contents = taskContentF.read().split()
    taskContentF.close()

    wordsCount = {}


    for word in file_contents:
        currentWord = word.lower()
        if( wordsCount.get(currentWord) == None ):
            wordsCount[currentWord] = 1
        else:
            wordsCount[currentWord] = wordsCount.get(currentWord) + 1

    f.seek(0)
    f.truncate(0)
    f.write(json.dumps(wordsCount))
    f.close()

process()