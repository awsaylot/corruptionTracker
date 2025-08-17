import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const RelationshipTypesPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Relationship Types Manager</h1>
        <p>Relationship type management interface will go here.</p>
      </div>
    </MainLayout>
  );
};

export default RelationshipTypesPage;
