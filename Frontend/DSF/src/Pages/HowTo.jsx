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

  return <Accordion collapseAll={true}>
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
}
