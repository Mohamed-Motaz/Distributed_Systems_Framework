import { Accordion } from 'flowbite-react'
import React from 'react'
import UploadFile from '../components/UploadFile.jsx'

export default function Home() {
  return <>
    <main className='pt-8'>

      <section className='m-8'>
        <UploadFile title='Process' />
        <UploadFile title='Distribute' />
        <UploadFile title='Aggregate' />

      </section>

    </main>
  </>
}
