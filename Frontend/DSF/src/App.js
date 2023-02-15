import React, { useEffect, useState } from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';

import RootLayout from './components/RootLayout';
import Home from './Pages/Home'
import NotFound from './Pages/NotFound'
import Landing from './Pages/Landing.jsx';
import HowTo from './Pages/HowTo.jsx';
import Manage from './Pages/Manage.jsx';
import SubmitJob from './Pages/SubmitJob.jsx';

import './App.css';
import Status from './Pages/Status.jsx';
import AboutUs from './Pages/AboutUs.jsx';





export default function App() 
{

  const [isFirst, setIsFirst] = useState(true)

  const HOME_ROUTE = createBrowserRouter([
    {
      path: '/', element: <RootLayout/>, children: [
        { index: true, element: isFirst ? <Landing/> : <Home /> },
        { path: '/how-to', element: <HowTo /> },
        { path: '/manage', element: <Manage /> },
        { path: '/submit-job', element: <SubmitJob /> },
        { path: '/status', element: <Status /> },
        { path: '/about-us', element: <AboutUs /> },
        // {
        //   path: '/movies', element: <Outlet></Outlet>,
        //   children: [
        //     { index: true, element: <Movies /> },
        //     { path: 'details/:type/:id', element: <ItemDetails /> },
        //   ]
        // },
        { path: '*', element: <NotFound /> },
      ]
    }
  ]);

  useEffect(() => {
    const result = localStorage.getItem('isFirst')
    if(result){
      setIsFirst(false)
      localStorage.setItem('isFirst','False')
    }
    else{
      localStorage.setItem('isFirst','True')
      setIsFirst(true)
    }

    console.log('====================================');
    console.log("DID APP MOUNT");
    console.log('====================================');
  }, [])

    return <RouterProvider router={HOME_ROUTE} />

}
