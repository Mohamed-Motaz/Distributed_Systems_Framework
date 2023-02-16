import { ProgressIndicator } from '@fluentui/react';
import { Progress, Tooltip } from 'flowbite-react';
import React, { useEffect, useState } from 'react'
import Loading from '../components/Loading.jsx';
import StatusCard from '../components/StatusCard.jsx';
import { WebSocketServerService } from '../services/WebSocketServerService.js';

function responseExample(status) {
  let response = {
    "response": {
      "Progress": [
        {
          "MasterId": "0c1f226b-089b-4462-9524-1abcd31da07d",
          "JobId": "35d42d8e-ee76-d4ff-fe2d-3102c15be475",
          "ClientId": "123",
          "Progress": 35,
          "Status": "Processing",
          "processBinary": 'Process.exe',
          "distributeBinary": 'Distribute.exe',
          "aggregateBinary": 'Aggregate.exe',
          "createdAt": '16-02-2023 12:00PM',
        }
      ],
      "error": false,
      "errorMsg": ""
    }
  }

  return status ? { ...response, "success": true } : { ...response, "error": "error message here" }
}




export default function Status() {

  const [jobs, setJobs] = useState(null)
  const [loading, setLoading] = useState(true)


  const getJobsProgress = async () => {
    //const jobProgress = await WebSocketServerService().getJobProgress();
    //setJobs(jobProgress || [])

    setJobs([responseExample(true), responseExample(false), responseExample(true)])
  };

  useEffect(() => {
    getJobsProgress()

    const intervalCalling = setInterval(async () => {
      //console.log("getJobsProgress() : Start...");
      await getJobsProgress()
      //console.log("getJobsProgress() : Done");
    }, 5000)

    return () => {
      clearInterval(intervalCalling)
    }
  }, [])


  return <main className='flex flex-col items-center pb-20 md:px-6'>
    <h1 className='md:text-5xl text-3xl mb-8'>
      Status
    </h1>

    {
      loading && jobs === null ?
        <Loading />
        : jobs.length ? <section className='w-full grid grid-cols-12 gap-8'>
          {
            jobs.map((job, index) => <StatusCard
              key={index}
              // key={job.response.Progress[0].JobId} 
              job={job}
            />)
          }
        </section>
          : <h2 className='text-2xl self-start'>No Jobs Found</h2>
    }


  </main>
}
