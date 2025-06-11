# Task Management Frontend (React + TypeScript)

A modern React frontend for the Task Management API built with TypeScript and Vite.

## Features

- ✅ Modern React 18 with TypeScript
- 🚀 Fast development with Vite
- 📱 Responsive design
- 🎨 Beautiful UI with CSS animations
- 🔄 Real-time task management (CRUD operations)
- 📊 Task statistics dashboard
- 🔍 Task filtering (All, Incomplete, Completed)
- 📄 Pagination support
- ⚙️ Configurable API port
- 🧪 Connection testing
- 📝 Edit tasks with modal interface

## Project Structure

```
src/
├── components/          # React components
│   ├── Header.tsx      # Header with stats and port config
│   ├── TaskForm.tsx    # Add new task form
│   ├── TaskFilter.tsx  # Filter buttons
│   ├── TaskList.tsx    # Task list container
│   ├── TaskItem.tsx    # Individual task item
│   ├── Pagination.tsx  # Pagination controls
│   ├── EditTaskModal.tsx # Edit task modal
│   └── Notification.tsx # Toast notifications
├── hooks/              # Custom React hooks
│   └── useTasks.ts    # Task management logic
├── services/           # API services
│   └── api.ts         # API communication
├── types/              # TypeScript type definitions
│   └── index.ts       # Shared types
├── styles/             # CSS styles
│   └── App.css        # Main stylesheet
├── App.tsx            # Main app component
└── main.tsx           # App entry point
```

## Prerequisites

- Node.js (v16 or higher)
- npm or yarn

## Installation

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

## Development

Start the development server:
```bash
npm run dev
```

The application will be available at `http://localhost:3666`

## Build

Build for production:
```bash
npm run build
```

Preview production build:
```bash
npm run preview
```

## Docker Deployment

### Build and Run with Docker

1. **Build the Docker image:**
   ```bash
   make build
   # or
   docker build -t task-frontend:latest .
   ```

2. **Run the Docker container:**
   ```bash
   make run
   # or
   docker run -d -p 3666:80 --name task-frontend task-frontend:latest
   ```

3. **Access the application:**
   - Local: `http://localhost:3666`
   - Network: `http://your-server-ip:3666`

### Docker Management Commands

```bash
# Build and deploy
make deploy

# View container status
make status

# View container logs
make logs

# Stop container
make stop

# Restart container
make restart

# Clean up
make clean
```

## Configuration

### API Port Configuration

The frontend can connect to different API ports:

1. **URL Parameter**: Add `?port=3666` to the URL
2. **UI Settings**: Use the port configuration in the header
3. **Local Storage**: The port setting is saved automatically

### Default Settings

- **Default API Port**: 3666
- **Default Frontend Port**: 8000
- **Tasks per page**: 10

## API Integration

The frontend integrates with the Task Management API with the following endpoints:

- `GET /api/v1/tasks/paginated` - Get paginated tasks
- `GET /api/v1/tasks/status/{status}` - Get tasks by status
- `POST /api/v1/tasks` - Create new task
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task
- `GET /api/v1/stats` - Get task statistics
- `GET /health` - Health check

## Components Overview

### Core Components

- **App**: Main application container
- **Header**: Displays statistics and port configuration
- **TaskForm**: Form for adding new tasks
- **TaskFilter**: Buttons for filtering tasks
- **TaskList**: Container for task items
- **TaskItem**: Individual task with actions
- **Pagination**: Navigation for multiple pages

### Utility Components

- **EditTaskModal**: Modal for editing tasks
- **Notification**: Toast notifications for user feedback

### Custom Hooks

- **useTasks**: Manages all task-related state and operations

## Features in Detail

### Task Management
- Create, read, update, and delete tasks
- Toggle task completion status
- Edit task name and status

### Filtering
- View all tasks
- Filter by incomplete tasks
- Filter by completed tasks

### Pagination
- Navigate through multiple pages of tasks
- Configurable items per page

### Responsive Design
- Mobile-friendly interface
- Adaptive layout for different screen sizes

### User Experience
- Loading states
- Error handling
- Success notifications
- Confirmation dialogs

## Technology Stack

- **React 18**: Modern React with hooks
- **TypeScript**: Type-safe development
- **Vite**: Fast build tool and dev server
- **CSS3**: Modern styling with animations
- **Fetch API**: HTTP client for API calls

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Development Notes

- All components are written in TypeScript
- Hooks are used for state management
- CSS uses modern features like backdrop-filter
- API calls are centralized in the service layer
- Error boundaries and proper error handling
- Accessible UI components