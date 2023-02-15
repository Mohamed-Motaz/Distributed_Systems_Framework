import { set } from 'lodash'
import React, { useContext, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AppContext } from '../context/AppContext.js'

export default function SubmitJob() {

  const navigate = useNavigate()
  const { apiEndPoint } = useContext(AppContext)

  const jobContentInput = useRef()

  const [isLoading, setIsLoading] = useState(false)
  const [job, setJob] = useState()

  const handleJobSubmit = async () => {
    setIsLoading(true)
    setJob({
      jobContent: jobContentInput.current.value,
      optionalFiles: '',
      process: '',
      distribute: '',
      aggregate: ''
    })

    try {
      
    } catch (error) {
      
    }

    setIsLoading(false)
    navigate('/status')
  }

  return <main className='flex flex-col items-center pb-20 md:px-16'>
    <h1 className='md:text-5xl text-3xl mb-8'>
      Submit Job
    </h1>

    <div className='flex flex-col shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full'>
      <section className='w-full flex items-center justify-start gap-2'>
        <h3 className='md:text-2xl text-xl '>
          API Endpoint
        </h3>
        <div className='w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1'>
          <p>
            {apiEndPoint || 'http://localhost:5000/api/v1/test'}
          </p>
        </div>
      </section>

      <section className='mt-6'>
        <h3 className='md:text-2xl text-xl '>
          Job Content
        </h3>
        <textarea className='w-full rounded-lg border-2 border-blue-800 outline-none bg-black' ref={jobContentInput} />
      </section>

      <section>
        put 4 buttions here:  <br />
        (1) add process       <br />
        (2)add distribute     <br />
        (3)add aggregate      <br />
        (4) upload optional files
      </section>

      <button className='rounded-lg px-14 py-2 bg-blue-800 w-fit mt-8 self-center text-xl' onClick={handleJobSubmit}>
        {isLoading? 'Submit...' : 'Submit Job' }
      </button>
      
    </div>
  </main>
}
