import json

filePath = "./aggregate.txt"

def aggregate():
    f = open(filePath, 'r+')
    finishedTasksPaths = f.read().replace("\n", "").strip().split(",")


    totalWordsCount = {}

    i = 0
    while i < len(finishedTasksPaths):
        taskF = open(finishedTasksPaths[i], 'r')

        wordsDict = taskF.read()
        wordsDict = json.loads(wordsDict)

        if i == 0:
            totalWordsCount = wordsDict
        else:
            for word in wordsDict:
                if totalWordsCount.get(word) is None:
                    totalWordsCount[word] = wordsDict.get(word)
                else:
                    totalWordsCount[word] = wordsDict.get(word) + totalWordsCount.get(word)

        taskF.close()
        i += 1

    f.seek(0)
    f.truncate(0)
    f.write(str(totalWordsCount))
    f.close()

aggregate()