import SimpleChat from '../../components/chat/SimpleChat';

export default function SimpleChatPage() {
  return (
    <div className="container mx-auto h-screen p-4">
      <div className="flex flex-col h-full max-w-4xl mx-auto">
        <h1 className="text-2xl font-bold mb-4">Simple Chat</h1>
        <div className="flex-1 bg-white rounded-lg shadow">
          <SimpleChat />
        </div>
      </div>
    </div>
  );
}
