import requests

def fetch_and_save_webpage(url, file_path):
    try:
        # Fetch the webpage
        response = requests.get(url)

        # Check if the request was successful (HTTP status code 200)
        if response.status_code == 200:
            # Save the webpage content to a file
            with open(file_path, 'w', encoding='utf-8') as file:
                file.write(response.text)
            print(f"Webpage saved successfully at {file_path}")
        else:
            print(f"Failed to fetch the webpage. Status code: {response.status_code}")
    except requests.RequestException as e:
        print(f"Error fetching the webpage: {e}")

# URL of the webpage to fetch
url = 'https://www.ri.se'

# Path to save the file
file_path = '/cfs/src/html'
#file_path = 'www.ri.se.html'

# Call the function
fetch_and_save_webpage(url, file_path)
