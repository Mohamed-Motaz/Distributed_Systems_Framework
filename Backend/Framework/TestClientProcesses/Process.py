import json

def process(taskFilePath):
    f = open(taskFilePath, 'r+')

    wordsCount = {}

    file_contents = f.read().split()

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
