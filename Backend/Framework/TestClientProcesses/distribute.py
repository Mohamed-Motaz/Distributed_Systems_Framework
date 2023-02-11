filePath = "./distribute.txt"

def distribute():
    f = open(filePath, 'r+')
    jobContent = f.read().replace("\n", ",").strip().split(",")

    for i in range(0, len(jobContent)):
        jobContent[i] = "./readingFiles/" + jobContent[i]
    
    print(jobContent)

    f.seek(0)
    f.truncate(0)
    f.write(",".join(jobContent))
    f.close()

distribute()