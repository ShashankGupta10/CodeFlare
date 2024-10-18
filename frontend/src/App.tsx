import { useState } from "react";

function App() {
  const [formData, setFormData] = useState({
    githubUrl: "",
    projectDir: "",
  });

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log(formData);
    const response = await fetch("http://localhost:8080/deploy", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        repo_url: formData.githubUrl,
        project_directory: formData.projectDir,
      }),
    });
    const data = await response.json();
    console.log(data);
  };
  return (
    <div className="h-screen flex justify-center items-center bg-gray-200">
      <div className="rounded-xl">
        <h1 className="text-2xl p-4 text-left text-black">Deploy React App</h1>
        <form onSubmit={handleSubmit}>
          <div className="flex flex-col gap-2">
            <label htmlFor="">Github URL</label>
            <input
              name="githubUrl"
              type="text"
              value={formData.githubUrl}
              onChange={onChange}
            />
          </div>
          <div className="flex flex-col gap-2">
            <label htmlFor="">Project Dir</label>
            <input
              type="text"
              name="projectDir"
              value={formData.projectDir}
              onChange={onChange}
            />
          </div>
          <button type="submit">Deploy</button>
        </form>
      </div>
    </div>
  );
}

export default App;
