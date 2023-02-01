import React from 'react'
import { Outlet } from 'react-router-dom'
import NavBar from './NavBar'


export default function RootLayout()
{

  return <>
    <NavBar/>
    <div className='dark pt-20 px-8'>
      <Outlet></Outlet>
    </div>
  </>
}
