import React, { useEffect, useState, useContext } from "react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import useWebSocket from "react-use-websocket";
import RootLayout from "./components/RootLayout";
import Home from "./Pages/Home";
import NotFound from "./Pages/NotFound";
import Landing from "./Pages/Landing.jsx";
import HowTo from "./Pages/HowTo.jsx";
import { AppContext } from "../src/context/AppContext";
import Manage from "./Pages/Manage.jsx";
import SubmitJob from "./Pages/SubmitJob.jsx";

import "./App.css";
import Status from "./Pages/Status.jsx";
import AboutUs from "./Pages/AboutUs.jsx";
import FinishedJobs from "./Pages/FinishedJobs.jsx";

export default function App() {
  const { clientId } = useContext(AppContext);
  const WS_URL = `ws://localhost:3001/openWS/${clientId}`;

  const wsClient = useWebSocket(WS_URL, {
    onOpen: () => {
      console.log("WebSocket connection established.");
    },
    onClose: () => {
      console.log(
        "WebSocket connection closed, it will be re-established in a second"
      );
      setTimeout(() => wsClient(), 1000);
    },
    onMessage: (e) => console.log({ e }),
  });

  const [isFirst, setIsFirst] = useState(true);

  const HOME_ROUTE = createBrowserRouter([
    {
      path: "/",
      element: <RootLayout />,
      children: [
        { index: true, element: isFirst ? <Landing /> : <Home /> },
        { path: "/how-to", element: <HowTo /> },
        { path: "/manage", element: <Manage /> },
        { path: "/submit-job", element: <SubmitJob wsClient={wsClient} /> },
        { path: "/status", element: <Status /> },
        { path: "/finished-jobs", element: <FinishedJobs /> },
        { path: "/about-us", element: <AboutUs /> },
        // {
        //   path: '/movies', element: <Outlet></Outlet>,
        //   children: [
        //     { index: true, element: <Movies /> },
        //     { path: 'details/:type/:id', element: <ItemDetails /> },
        //   ]
        // },
        { path: "*", element: <NotFound /> },
      ],
    },
  ]);

  useEffect(() => {
    const result = localStorage.getItem("isFirst");
    if (result) {
      setIsFirst(false);
      localStorage.setItem("isFirst", "False");
    } else {
      localStorage.setItem("isFirst", "True");
      setIsFirst(true);
    }

    console.log("====================================");
    console.log("DID APP MOUNT");
    console.log("====================================");
  }, []);

  return <RouterProvider router={HOME_ROUTE} />;
}
