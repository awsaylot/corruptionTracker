import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const RelationshipsPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Relationships Overview</h1>
        <p>Relationship management interface will go here.</p>
      </div>
    </MainLayout>
  );
};

export default RelationshipsPage;
