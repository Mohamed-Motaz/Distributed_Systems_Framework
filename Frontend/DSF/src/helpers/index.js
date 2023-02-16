import {saveAs} from 'file-saver'

export async function downloadImage(item, nameToSaveWith) {
    saveAs(item, nameToSaveWith)
}