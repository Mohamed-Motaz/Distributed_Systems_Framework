import React from 'react'

function statusCard(job) {
  return <div className='xl:col-span-4 md:col-span-6 col-span-12 shadow-card hover:shadow-cardhover rounded-lg p-8 gap-2'>
    <h2 className='text-xl text-center mb-5'>
      {job.jobName}
    </h2>
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
    <div className='w-full h-5 rounded-lg bg-gray-600'>
      <div className={`w-[${job.progress}] h-5 rounded-lg bg-green-600`}>
      </div>
    </div>
  </div>
}

export default function Status() {

  const list = [
    {
      jobName: 'Job One',
      processBinary: 'Process.exe',
      distributeBinary: 'Distribute.exe',
      aggregateBinary: 'Aggregate.exe',
      createdAt: '16-02-2023 12:00PM',
      progress: '80%'
    },
    {
      jobName: 'Job Two',
      processBinary: 'Process.exe',
      distributeBinary: 'Distribute.exe',
      aggregateBinary: 'Aggregate.exe',
      createdAt: '16-02-2023 12:00PM',
      progress: '60%'
    },
    {
      jobName: 'Job Three',
      processBinary: 'Process.exe',
      distributeBinary: 'Distribute.exe',
      aggregateBinary: 'Aggregate.exe',
      createdAt: '16-02-2023 12:00PM',
      progress: '25%'
    },
  ]

  return <main className='flex flex-col items-center pb-20 md:px-6'>
    <h1 className='md:text-5xl text-3xl mb-8'>
      Status
    </h1>

    <section className='w-full grid grid-cols-12 gap-8'>
      {
        list.map(job => statusCard(job))
      }
    </section>

  </main>
}
