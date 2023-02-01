import React from 'react'
import UploadFileButton from '../components/UploadFileButton.jsx'

export default function Home() {
  return <>
    <main className='pt-8'>

      <section className='m-8'>
        <UploadFileButton title='Process' />
        <UploadFileButton title='Distribute' />
        <UploadFileButton title='Aggregate' />

      </section>

    </main>
  </>
}
