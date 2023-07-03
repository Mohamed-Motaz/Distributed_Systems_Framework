import { Modal } from "flowbite-react";
import React, { useContext, useState } from "react";
import { AppContext } from "../context/AppContext";

export default function WorkersModal({ workers }) {
  console.log({ workers });
  const [isOpen, setIsOpen] = useState(false);
  const toggleModal = () => {
    setIsOpen(!isOpen);
  };
  const { clientId } = useContext(AppContext);

  return (
    <section className="w-full flex justify-center pt-5">
      <button
        onClick={toggleModal}
        className="rounded-lg px-10 py-2 bg-blue-800"
      >
        Show Workers
      </button>
      <Modal show={isOpen} onClose={toggleModal} className="dark">
        <Modal.Header className="w-full bg-[#141414] py-4">
          Workers Servers
        </Modal.Header>

        <section className="w-full grid grid-cols-12 max-h-[500px] overflow-y-auto bg-[#141414] py-4">
          {workers?.map((task) => (
            <div className="col-span-12 px-3 mb-6 flex">
              <div className="rounded-lg p-8 self-stretch w-full gap-2 bg-green-900">
                <div className="mb-3 flex items-center gap-2">
                  <div className="w-24">Worker ID:</div>
                  <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
                    <p>{task.WorkerId}</p>
                  </div>
                </div>
                <div className="mb-3 flex items-center gap-2">
                  <div className="w-24">Client ID:</div>
                  <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
                    <p>{clientId}</p>
                  </div>
                </div>
                <div className="mb-3 flex items-center gap-2">
                  <div className="w-24">Current Task:</div>
                  <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
                    <p>{task.CurrentTaskContent}</p>
                  </div>
                </div>
                <div className="mb-3 flex items-start gap-2">
                  <div className="w-24">Finished Tasks:</div>
                  <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
                    <p>
                      {task.FinishedTasksContent?.map((content) => (
                        <p> {content} </p>
                      ))}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </section>
      </Modal>
    </section>
  );
}
