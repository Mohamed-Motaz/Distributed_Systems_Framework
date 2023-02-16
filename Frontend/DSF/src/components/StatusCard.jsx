import { Progress, Tooltip } from 'flowbite-react';
import React, { useState } from 'react'

export default function StatusCard({ job }) {
    console.log(job.error? 'Error Found': 'Success');
    job = job.response.Progress[0];


    const [isCopied, setIsCopied] = useState(false)
    const [timer, setTimer] = useState(null)

    function copyToClipboard() {
        clearTimeout(timer)
        navigator.clipboard.writeText(job.JobId)
        setIsCopied(true)
        setTimer(setTimeout(() => setIsCopied(false), 1000))
    }

    return <div className='xl:col-span-4 md:col-span-6 col-span-12 shadow-card hover:shadow-cardhover rounded-lg p-8 gap-2'>
        <div className='mb-3 flex justify-center items-center gap-2'>
            <button className='rounded-lg bg-blue-800' onClick={copyToClipboard}>
                <Tooltip content={isCopied ? 'Copied To Cilpboard' : job.JobId}>
                    <h2 className='text-xl text-center px-6 py-2' >
                        JOB ID
                    </h2>
                </Tooltip>
            </button>
        </div>

        <div className='mb-3 flex items-center gap-2'>
            Process:
            <div className='w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1'>
                <p>{job.processBinary}</p>
            </div>
        </div>
        <div className='mb-3 flex items-center gap-2'>
            Distribute:
            <div className='w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1'>
                <p>{job.distributeBinary}</p>
            </div>
        </div>
        <div className='mb-3 flex items-center gap-2'>
            Aggregate:
            <div className='w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1'>
                <p>{job.aggregateBinary}</p>
            </div>
        </div>
        <div className='mb-3 flex items-center gap-2'>
            Created At:
            <div className='w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1'>
                <p>{job.createdAt}</p>
            </div>
        </div>
        <div className='mt-6 mb-3'>
            <Progress
                progress={job.Progress}
                color="green"
                label={job.Status}
                labelPosition="outside"
                labelProgress={true}
            />
        </div>
    </div>
}
