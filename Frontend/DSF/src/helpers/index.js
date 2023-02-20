import {saveAs} from 'file-saver'

export async function downloadItem(item, nameToSaveWith) {
    saveAs(item, nameToSaveWith)
}