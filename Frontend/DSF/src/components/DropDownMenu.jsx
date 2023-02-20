import { Dropdown } from 'flowbite-react'
import { NavLink } from 'react-router-dom';


export default function DropDownMenu() {


    return <Dropdown className='dark' label="Menu" style={{ background: '#1744e1' }}>

        <NavLink className="nav-link" to='/'>
            <Dropdown.Item className='justify-center'>
                Home
            </Dropdown.Item>

        </NavLink>

        <NavLink className="nav-link" to='/how-to'>
            <Dropdown.Item className='justify-center'>
                How To
            </Dropdown.Item>
        </NavLink >
        <NavLink className="nav-link" to='/manage'>
            <Dropdown.Item className='justify-center'>
                Manage
            </Dropdown.Item>
        </NavLink>
        <NavLink className="nav-link" to='/submit-job'>
            <Dropdown.Item className='justify-center'>
                Submit Job
            </Dropdown.Item>
        </NavLink>
        <NavLink className="nav-link" to='/status'>
            <Dropdown.Item className='justify-center'>
                Status
            </Dropdown.Item>
        </NavLink>
        <NavLink className="nav-link" to='/finished-jobs'>
            <Dropdown.Item className='justify-center'>
                Finished
            </Dropdown.Item>
        </NavLink>
        <NavLink className="nav-link" to='/about-us'>
            <Dropdown.Item className='justify-center'>
                About Us
            </Dropdown.Item>
        </NavLink>
    </Dropdown>
}
