import React from 'react'
import { downloadItem } from '../helpers/index.js'

export default function FinishedJobs() {


    const list = [
        {
            jobId: "8jd83-93nKJds9-2809qr00dd",
            attr1: "Attribute One",
            attr2: "Attribute Two",
        },
        {
            jobId: "8jd83-93nKJds9-2809qr00dd",
            attr1: "Attribute One",
            attr2: "Attribute Two",
        },
        {
            jobId: "8jd83-93nKJds9-2809qr00dd",
            attr1: "Attribute One",
            attr2: "Attribute Two",
        },
    ]



    return <main className='flex flex-col items-center pb-20 md:px-6'>
        <h1 className='md:text-5xl text-3xl mb-16'>
            Finished Jobs
        </h1>
        <section>
            <table class="w-full table-fixed">
                <thead>
                    <tr>
                        <th className='pb-5 text-xl'>Job ID</th>
                        <th className='pb-5 text-xl'>Attribute</th>
                        <th className='pb-5 text-xl'>Attribute</th>
                        <th className='pb-5 text-xl'>Download</th>
                    </tr>
                </thead>
                <tbody>
                    {
                        list.map((row, i) => <tr className={`border-t-2 border-b-2 ${i % 2 === 0 && 'bg-white bg-opacity-10'}`}>
                            <td className='text-center'>{row.jobId}</td>
                            <td className='text-center'>{row.attr1}</td>
                            <td className='text-center'>{row.attr2}</td>
                            <td className='text-center'>
                                <button 
                                className="my-2 px-4 py-2 rounded-lg bg-blue-800" 
                                onClick={() => { downloadItem(row,`${row.jobId}.txt`) }}>
                                    DOWNLOAD
                                </button>
                            </td>
                        </tr>)
                    }
                </tbody>
            </table>
        </section>
    </main>
}
