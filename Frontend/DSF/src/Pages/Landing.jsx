import React, { useContext, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { AppContext } from '../context/AppContext.js'

export default function Landing() {

  const navigate = useNavigate();
  const { changeApiEndPoint } = useContext(AppContext)
  const apiEndPointInput = useRef()

  const handleOnClick = () => {
    changeApiEndPoint(apiEndPointInput.current.value)
    navigate('/how-to')
  }

  return <main className='flex flex-col items-center'>
    <h1 className='md:text-5xl text-3xl mb-16'>
      Distributed Systems Framework
    </h1>

    <div className='flex flex-col justify-center items-center shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2 max-w-xl w-full'>
      <h3 className='md:text-2xl text-xl'>
        Enter API Endpoint
      </h3>
      <input className='w-full rounded-lg border-2 border-blue-800 outline-none bg-black' type="text" ref={apiEndPointInput} />
      <button className='rounded-lg px-10 py-2 bg-blue-800 mt-10' onClick={handleOnClick}>
        GO
      </button>
    </div>
  </main>
}
