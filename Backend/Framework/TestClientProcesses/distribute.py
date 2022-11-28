filePath = "./distribute.txt"

def distribute():
    f = open(filePath, 'r+')
    jobContentFilePath = f.read().replace("\n", "").strip()


    jobContentF = open(jobContentFilePath, "r")
    tasks = []
    file_contents = jobContentF.readlines()
    jobContentF.close()

    
    for taskString in file_contents:
        tasks.append(taskString.replace('\n', ''))

    f.seek(0)
    f.truncate(0)
    f.write(str(tasks))
    f.close()

distribute()