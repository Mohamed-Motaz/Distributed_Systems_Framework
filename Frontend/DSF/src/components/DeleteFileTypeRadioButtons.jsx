import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";

export const DeleteFileTypeRadioButtons = (props) => {
  const { deleteFileType, setDeleteFileType } = props;

  const handleOnChagne = (type) => setDeleteFileType(type);
  
  return (
    <div className="flex">
      <div className="flex items-center mr-4">
        <input
          id="inline-delete-2-radio"
          type="radio"
          value=""
          onChange={() => handleOnChagne(BinariesType.Distribute)}
          name="inline-delete-radio-group"
          className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
        />
        <label
          htmlFor="inline-delete-2-radio"
          className="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300"
        >
          Distribute
        </label>
      </div>
      <div className="flex items-center mr-4">
        <input
          id="inline-delete-radio"
          type="radio"
          value=""
          onChange={() => handleOnChagne(BinariesType.process)}
          name="inline-delete-radio-group"
          className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
        />
        <label
          htmlFor="inline-delete-radio"
          className="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300"
        >
          Process
        </label>
      </div>
      <div className="flex items-center mr-4">
        <input
          id="inline-delete-checked-radio"
          onChange={() => handleOnChagne(BinariesType.aggregate)}
          type="radio"
          value=""
          name="inline-delete-radio-group"
          className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
        />
        <label
          htmlFor="inline-delete-checked-radio"
          className="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300"
        >
          Aggregate
        </label>
      </div>
    </div>
  );
};
