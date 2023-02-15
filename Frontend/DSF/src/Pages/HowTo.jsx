import { Accordion } from 'flowbite-react'
import React from 'react'

export default function HowTo() {

  const StepsList = [
    {
      title: "Step 1",
      content: "Lorem Ipsum is simply dummy text."
    },
    {
      title: "Step 2",
      content: "Lorem Ipsum is simply dummy text."
    },
    {
      title: "Step 3",
      content: "Lorem Ipsum is simply dummy text."
    },
    {
      title: "Step 4",
      content: "Lorem Ipsum is simply dummy text."
    },
    {
      title: "Step 5",
      content: "Lorem Ipsum is simply dummy text."
    },
  ]

  return <main className='w-full flex flex-col items-stretch'>
    <h1 className='md:text-5xl text-3xl mb-8 self-center'>
      How To Use
    </h1>
    <Accordion collapseAll={true}>
      {
        StepsList.map(item => <Accordion.Panel key={item.title}>
          <Accordion.Title>
            {item.title}
          </Accordion.Title>
          <Accordion.Content>
            <p className="mb-2 text-gray-500 dark:text-gray-400">
              {item.content}
            </p>
          </Accordion.Content>
        </Accordion.Panel>)
      }
    </Accordion>
  </main>
}
