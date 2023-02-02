import React from 'react'
import '../css/LoadingStyle.css'

export default function Loading() {
  return<section id='loading' className='mt-4 mb-3 w-full flex justify-center items-center'>
    <div className='fa-spin spin-fast'></div>
  </section>
}
