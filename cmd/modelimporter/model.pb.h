// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: model.proto

#ifndef PROTOBUF_model_2eproto__INCLUDED
#define PROTOBUF_model_2eproto__INCLUDED

#include <string>

#include <google/protobuf/stubs/common.h>

#if GOOGLE_PROTOBUF_VERSION < 3001000
#error This file was generated by a newer version of protoc which is
#error incompatible with your Protocol Buffer headers.  Please update
#error your headers.
#endif
#if 3001000 < GOOGLE_PROTOBUF_MIN_PROTOC_VERSION
#error This file was generated by an older version of protoc which is
#error incompatible with your Protocol Buffer headers.  Please
#error regenerate this file with a newer version of protoc.
#endif

#include <google/protobuf/arena.h>
#include <google/protobuf/arenastring.h>
#include <google/protobuf/generated_message_util.h>
#include <google/protobuf/metadata.h>
#include <google/protobuf/message.h>
#include <google/protobuf/repeated_field.h>
#include <google/protobuf/extension_set.h>
#include <google/protobuf/unknown_field_set.h>
// @@protoc_insertion_point(includes)

namespace protos {

// Internal implementation detail -- do not call these.
void protobuf_AddDesc_model_2eproto();
void protobuf_InitDefaults_model_2eproto();
void protobuf_AssignDesc_model_2eproto();
void protobuf_ShutdownFile_model_2eproto();

class Mesh;
class Model;

// ===================================================================

class Mesh : public ::google::protobuf::Message /* @@protoc_insertion_point(class_definition:protos.Mesh) */ {
 public:
  Mesh();
  virtual ~Mesh();

  Mesh(const Mesh& from);

  inline Mesh& operator=(const Mesh& from) {
    CopyFrom(from);
    return *this;
  }

  static const ::google::protobuf::Descriptor* descriptor();
  static const Mesh& default_instance();

  static const Mesh* internal_default_instance();

  void Swap(Mesh* other);

  // implements Message ----------------------------------------------

  inline Mesh* New() const { return New(NULL); }

  Mesh* New(::google::protobuf::Arena* arena) const;
  void CopyFrom(const ::google::protobuf::Message& from);
  void MergeFrom(const ::google::protobuf::Message& from);
  void CopyFrom(const Mesh& from);
  void MergeFrom(const Mesh& from);
  void Clear();
  bool IsInitialized() const;

  size_t ByteSizeLong() const;
  bool MergePartialFromCodedStream(
      ::google::protobuf::io::CodedInputStream* input);
  void SerializeWithCachedSizes(
      ::google::protobuf::io::CodedOutputStream* output) const;
  ::google::protobuf::uint8* InternalSerializeWithCachedSizesToArray(
      bool deterministic, ::google::protobuf::uint8* output) const;
  ::google::protobuf::uint8* SerializeWithCachedSizesToArray(::google::protobuf::uint8* output) const {
    return InternalSerializeWithCachedSizesToArray(false, output);
  }
  int GetCachedSize() const { return _cached_size_; }
  private:
  void SharedCtor();
  void SharedDtor();
  void SetCachedSize(int size) const;
  void InternalSwap(Mesh* other);
  void UnsafeMergeFrom(const Mesh& from);
  private:
  inline ::google::protobuf::Arena* GetArenaNoVirtual() const {
    return _internal_metadata_.arena();
  }
  inline void* MaybeArenaPtr() const {
    return _internal_metadata_.raw_arena_ptr();
  }
  public:

  ::google::protobuf::Metadata GetMetadata() const;

  // nested types ----------------------------------------------------

  // accessors -------------------------------------------------------

  // optional bytes indices = 1;
  void clear_indices();
  static const int kIndicesFieldNumber = 1;
  const ::std::string& indices() const;
  void set_indices(const ::std::string& value);
  void set_indices(const char* value);
  void set_indices(const void* value, size_t size);
  ::std::string* mutable_indices();
  ::std::string* release_indices();
  void set_allocated_indices(::std::string* indices);

  // optional bytes positions = 2;
  void clear_positions();
  static const int kPositionsFieldNumber = 2;
  const ::std::string& positions() const;
  void set_positions(const ::std::string& value);
  void set_positions(const char* value);
  void set_positions(const void* value, size_t size);
  ::std::string* mutable_positions();
  ::std::string* release_positions();
  void set_allocated_positions(::std::string* positions);

  // optional bytes normals = 3;
  void clear_normals();
  static const int kNormalsFieldNumber = 3;
  const ::std::string& normals() const;
  void set_normals(const ::std::string& value);
  void set_normals(const char* value);
  void set_normals(const void* value, size_t size);
  ::std::string* mutable_normals();
  ::std::string* release_normals();
  void set_allocated_normals(::std::string* normals);

  // optional bytes tangents = 4;
  void clear_tangents();
  static const int kTangentsFieldNumber = 4;
  const ::std::string& tangents() const;
  void set_tangents(const ::std::string& value);
  void set_tangents(const char* value);
  void set_tangents(const void* value, size_t size);
  ::std::string* mutable_tangents();
  ::std::string* release_tangents();
  void set_allocated_tangents(::std::string* tangents);

  // optional bytes bitangents = 5;
  void clear_bitangents();
  static const int kBitangentsFieldNumber = 5;
  const ::std::string& bitangents() const;
  void set_bitangents(const ::std::string& value);
  void set_bitangents(const char* value);
  void set_bitangents(const void* value, size_t size);
  ::std::string* mutable_bitangents();
  ::std::string* release_bitangents();
  void set_allocated_bitangents(::std::string* bitangents);

  // optional bytes tcoords = 6;
  void clear_tcoords();
  static const int kTcoordsFieldNumber = 6;
  const ::std::string& tcoords() const;
  void set_tcoords(const ::std::string& value);
  void set_tcoords(const char* value);
  void set_tcoords(const void* value, size_t size);
  ::std::string* mutable_tcoords();
  ::std::string* release_tcoords();
  void set_allocated_tcoords(::std::string* tcoords);

  // optional bytes albedo_map = 7;
  void clear_albedo_map();
  static const int kAlbedoMapFieldNumber = 7;
  const ::std::string& albedo_map() const;
  void set_albedo_map(const ::std::string& value);
  void set_albedo_map(const char* value);
  void set_albedo_map(const void* value, size_t size);
  ::std::string* mutable_albedo_map();
  ::std::string* release_albedo_map();
  void set_allocated_albedo_map(::std::string* albedo_map);

  // optional bytes normal_map = 8;
  void clear_normal_map();
  static const int kNormalMapFieldNumber = 8;
  const ::std::string& normal_map() const;
  void set_normal_map(const ::std::string& value);
  void set_normal_map(const char* value);
  void set_normal_map(const void* value, size_t size);
  ::std::string* mutable_normal_map();
  ::std::string* release_normal_map();
  void set_allocated_normal_map(::std::string* normal_map);

  // optional bytes rough_map = 9;
  void clear_rough_map();
  static const int kRoughMapFieldNumber = 9;
  const ::std::string& rough_map() const;
  void set_rough_map(const ::std::string& value);
  void set_rough_map(const char* value);
  void set_rough_map(const void* value, size_t size);
  ::std::string* mutable_rough_map();
  ::std::string* release_rough_map();
  void set_allocated_rough_map(::std::string* rough_map);

  // optional bytes metal_map = 10;
  void clear_metal_map();
  static const int kMetalMapFieldNumber = 10;
  const ::std::string& metal_map() const;
  void set_metal_map(const ::std::string& value);
  void set_metal_map(const char* value);
  void set_metal_map(const void* value, size_t size);
  ::std::string* mutable_metal_map();
  ::std::string* release_metal_map();
  void set_allocated_metal_map(::std::string* metal_map);

  // optional string state = 11;
  void clear_state();
  static const int kStateFieldNumber = 11;
  const ::std::string& state() const;
  void set_state(const ::std::string& value);
  void set_state(const char* value);
  void set_state(const char* value, size_t size);
  ::std::string* mutable_state();
  ::std::string* release_state();
  void set_allocated_state(::std::string* state);

  // optional string name = 12;
  void clear_name();
  static const int kNameFieldNumber = 12;
  const ::std::string& name() const;
  void set_name(const ::std::string& value);
  void set_name(const char* value);
  void set_name(const char* value, size_t size);
  ::std::string* mutable_name();
  ::std::string* release_name();
  void set_allocated_name(::std::string* name);

  // @@protoc_insertion_point(class_scope:protos.Mesh)
 private:

  ::google::protobuf::internal::InternalMetadataWithArena _internal_metadata_;
  ::google::protobuf::internal::ArenaStringPtr indices_;
  ::google::protobuf::internal::ArenaStringPtr positions_;
  ::google::protobuf::internal::ArenaStringPtr normals_;
  ::google::protobuf::internal::ArenaStringPtr tangents_;
  ::google::protobuf::internal::ArenaStringPtr bitangents_;
  ::google::protobuf::internal::ArenaStringPtr tcoords_;
  ::google::protobuf::internal::ArenaStringPtr albedo_map_;
  ::google::protobuf::internal::ArenaStringPtr normal_map_;
  ::google::protobuf::internal::ArenaStringPtr rough_map_;
  ::google::protobuf::internal::ArenaStringPtr metal_map_;
  ::google::protobuf::internal::ArenaStringPtr state_;
  ::google::protobuf::internal::ArenaStringPtr name_;
  mutable int _cached_size_;
  friend void  protobuf_InitDefaults_model_2eproto_impl();
  friend void  protobuf_AddDesc_model_2eproto_impl();
  friend void protobuf_AssignDesc_model_2eproto();
  friend void protobuf_ShutdownFile_model_2eproto();

  void InitAsDefaultInstance();
};
extern ::google::protobuf::internal::ExplicitlyConstructed<Mesh> Mesh_default_instance_;

// -------------------------------------------------------------------

class Model : public ::google::protobuf::Message /* @@protoc_insertion_point(class_definition:protos.Model) */ {
 public:
  Model();
  virtual ~Model();

  Model(const Model& from);

  inline Model& operator=(const Model& from) {
    CopyFrom(from);
    return *this;
  }

  static const ::google::protobuf::Descriptor* descriptor();
  static const Model& default_instance();

  static const Model* internal_default_instance();

  void Swap(Model* other);

  // implements Message ----------------------------------------------

  inline Model* New() const { return New(NULL); }

  Model* New(::google::protobuf::Arena* arena) const;
  void CopyFrom(const ::google::protobuf::Message& from);
  void MergeFrom(const ::google::protobuf::Message& from);
  void CopyFrom(const Model& from);
  void MergeFrom(const Model& from);
  void Clear();
  bool IsInitialized() const;

  size_t ByteSizeLong() const;
  bool MergePartialFromCodedStream(
      ::google::protobuf::io::CodedInputStream* input);
  void SerializeWithCachedSizes(
      ::google::protobuf::io::CodedOutputStream* output) const;
  ::google::protobuf::uint8* InternalSerializeWithCachedSizesToArray(
      bool deterministic, ::google::protobuf::uint8* output) const;
  ::google::protobuf::uint8* SerializeWithCachedSizesToArray(::google::protobuf::uint8* output) const {
    return InternalSerializeWithCachedSizesToArray(false, output);
  }
  int GetCachedSize() const { return _cached_size_; }
  private:
  void SharedCtor();
  void SharedDtor();
  void SetCachedSize(int size) const;
  void InternalSwap(Model* other);
  void UnsafeMergeFrom(const Model& from);
  private:
  inline ::google::protobuf::Arena* GetArenaNoVirtual() const {
    return _internal_metadata_.arena();
  }
  inline void* MaybeArenaPtr() const {
    return _internal_metadata_.raw_arena_ptr();
  }
  public:

  ::google::protobuf::Metadata GetMetadata() const;

  // nested types ----------------------------------------------------

  // accessors -------------------------------------------------------

  // repeated .protos.Mesh meshes = 1;
  int meshes_size() const;
  void clear_meshes();
  static const int kMeshesFieldNumber = 1;
  const ::protos::Mesh& meshes(int index) const;
  ::protos::Mesh* mutable_meshes(int index);
  ::protos::Mesh* add_meshes();
  ::google::protobuf::RepeatedPtrField< ::protos::Mesh >*
      mutable_meshes();
  const ::google::protobuf::RepeatedPtrField< ::protos::Mesh >&
      meshes() const;

  // @@protoc_insertion_point(class_scope:protos.Model)
 private:

  ::google::protobuf::internal::InternalMetadataWithArena _internal_metadata_;
  ::google::protobuf::RepeatedPtrField< ::protos::Mesh > meshes_;
  mutable int _cached_size_;
  friend void  protobuf_InitDefaults_model_2eproto_impl();
  friend void  protobuf_AddDesc_model_2eproto_impl();
  friend void protobuf_AssignDesc_model_2eproto();
  friend void protobuf_ShutdownFile_model_2eproto();

  void InitAsDefaultInstance();
};
extern ::google::protobuf::internal::ExplicitlyConstructed<Model> Model_default_instance_;

// ===================================================================


// ===================================================================

#if !PROTOBUF_INLINE_NOT_IN_HEADERS
// Mesh

// optional bytes indices = 1;
inline void Mesh::clear_indices() {
  indices_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::indices() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.indices)
  return indices_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_indices(const ::std::string& value) {
  
  indices_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.indices)
}
inline void Mesh::set_indices(const char* value) {
  
  indices_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.indices)
}
inline void Mesh::set_indices(const void* value, size_t size) {
  
  indices_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.indices)
}
inline ::std::string* Mesh::mutable_indices() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.indices)
  return indices_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_indices() {
  // @@protoc_insertion_point(field_release:protos.Mesh.indices)
  
  return indices_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_indices(::std::string* indices) {
  if (indices != NULL) {
    
  } else {
    
  }
  indices_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), indices);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.indices)
}

// optional bytes positions = 2;
inline void Mesh::clear_positions() {
  positions_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::positions() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.positions)
  return positions_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_positions(const ::std::string& value) {
  
  positions_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.positions)
}
inline void Mesh::set_positions(const char* value) {
  
  positions_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.positions)
}
inline void Mesh::set_positions(const void* value, size_t size) {
  
  positions_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.positions)
}
inline ::std::string* Mesh::mutable_positions() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.positions)
  return positions_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_positions() {
  // @@protoc_insertion_point(field_release:protos.Mesh.positions)
  
  return positions_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_positions(::std::string* positions) {
  if (positions != NULL) {
    
  } else {
    
  }
  positions_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), positions);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.positions)
}

// optional bytes normals = 3;
inline void Mesh::clear_normals() {
  normals_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::normals() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.normals)
  return normals_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_normals(const ::std::string& value) {
  
  normals_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.normals)
}
inline void Mesh::set_normals(const char* value) {
  
  normals_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.normals)
}
inline void Mesh::set_normals(const void* value, size_t size) {
  
  normals_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.normals)
}
inline ::std::string* Mesh::mutable_normals() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.normals)
  return normals_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_normals() {
  // @@protoc_insertion_point(field_release:protos.Mesh.normals)
  
  return normals_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_normals(::std::string* normals) {
  if (normals != NULL) {
    
  } else {
    
  }
  normals_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), normals);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.normals)
}

// optional bytes tangents = 4;
inline void Mesh::clear_tangents() {
  tangents_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::tangents() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.tangents)
  return tangents_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_tangents(const ::std::string& value) {
  
  tangents_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.tangents)
}
inline void Mesh::set_tangents(const char* value) {
  
  tangents_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.tangents)
}
inline void Mesh::set_tangents(const void* value, size_t size) {
  
  tangents_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.tangents)
}
inline ::std::string* Mesh::mutable_tangents() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.tangents)
  return tangents_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_tangents() {
  // @@protoc_insertion_point(field_release:protos.Mesh.tangents)
  
  return tangents_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_tangents(::std::string* tangents) {
  if (tangents != NULL) {
    
  } else {
    
  }
  tangents_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), tangents);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.tangents)
}

// optional bytes bitangents = 5;
inline void Mesh::clear_bitangents() {
  bitangents_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::bitangents() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.bitangents)
  return bitangents_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_bitangents(const ::std::string& value) {
  
  bitangents_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.bitangents)
}
inline void Mesh::set_bitangents(const char* value) {
  
  bitangents_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.bitangents)
}
inline void Mesh::set_bitangents(const void* value, size_t size) {
  
  bitangents_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.bitangents)
}
inline ::std::string* Mesh::mutable_bitangents() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.bitangents)
  return bitangents_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_bitangents() {
  // @@protoc_insertion_point(field_release:protos.Mesh.bitangents)
  
  return bitangents_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_bitangents(::std::string* bitangents) {
  if (bitangents != NULL) {
    
  } else {
    
  }
  bitangents_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), bitangents);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.bitangents)
}

// optional bytes tcoords = 6;
inline void Mesh::clear_tcoords() {
  tcoords_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::tcoords() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.tcoords)
  return tcoords_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_tcoords(const ::std::string& value) {
  
  tcoords_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.tcoords)
}
inline void Mesh::set_tcoords(const char* value) {
  
  tcoords_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.tcoords)
}
inline void Mesh::set_tcoords(const void* value, size_t size) {
  
  tcoords_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.tcoords)
}
inline ::std::string* Mesh::mutable_tcoords() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.tcoords)
  return tcoords_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_tcoords() {
  // @@protoc_insertion_point(field_release:protos.Mesh.tcoords)
  
  return tcoords_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_tcoords(::std::string* tcoords) {
  if (tcoords != NULL) {
    
  } else {
    
  }
  tcoords_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), tcoords);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.tcoords)
}

// optional bytes albedo_map = 7;
inline void Mesh::clear_albedo_map() {
  albedo_map_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::albedo_map() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.albedo_map)
  return albedo_map_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_albedo_map(const ::std::string& value) {
  
  albedo_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.albedo_map)
}
inline void Mesh::set_albedo_map(const char* value) {
  
  albedo_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.albedo_map)
}
inline void Mesh::set_albedo_map(const void* value, size_t size) {
  
  albedo_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.albedo_map)
}
inline ::std::string* Mesh::mutable_albedo_map() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.albedo_map)
  return albedo_map_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_albedo_map() {
  // @@protoc_insertion_point(field_release:protos.Mesh.albedo_map)
  
  return albedo_map_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_albedo_map(::std::string* albedo_map) {
  if (albedo_map != NULL) {
    
  } else {
    
  }
  albedo_map_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), albedo_map);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.albedo_map)
}

// optional bytes normal_map = 8;
inline void Mesh::clear_normal_map() {
  normal_map_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::normal_map() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.normal_map)
  return normal_map_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_normal_map(const ::std::string& value) {
  
  normal_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.normal_map)
}
inline void Mesh::set_normal_map(const char* value) {
  
  normal_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.normal_map)
}
inline void Mesh::set_normal_map(const void* value, size_t size) {
  
  normal_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.normal_map)
}
inline ::std::string* Mesh::mutable_normal_map() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.normal_map)
  return normal_map_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_normal_map() {
  // @@protoc_insertion_point(field_release:protos.Mesh.normal_map)
  
  return normal_map_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_normal_map(::std::string* normal_map) {
  if (normal_map != NULL) {
    
  } else {
    
  }
  normal_map_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), normal_map);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.normal_map)
}

// optional bytes rough_map = 9;
inline void Mesh::clear_rough_map() {
  rough_map_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::rough_map() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.rough_map)
  return rough_map_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_rough_map(const ::std::string& value) {
  
  rough_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.rough_map)
}
inline void Mesh::set_rough_map(const char* value) {
  
  rough_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.rough_map)
}
inline void Mesh::set_rough_map(const void* value, size_t size) {
  
  rough_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.rough_map)
}
inline ::std::string* Mesh::mutable_rough_map() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.rough_map)
  return rough_map_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_rough_map() {
  // @@protoc_insertion_point(field_release:protos.Mesh.rough_map)
  
  return rough_map_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_rough_map(::std::string* rough_map) {
  if (rough_map != NULL) {
    
  } else {
    
  }
  rough_map_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), rough_map);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.rough_map)
}

// optional bytes metal_map = 10;
inline void Mesh::clear_metal_map() {
  metal_map_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::metal_map() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.metal_map)
  return metal_map_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_metal_map(const ::std::string& value) {
  
  metal_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.metal_map)
}
inline void Mesh::set_metal_map(const char* value) {
  
  metal_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.metal_map)
}
inline void Mesh::set_metal_map(const void* value, size_t size) {
  
  metal_map_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.metal_map)
}
inline ::std::string* Mesh::mutable_metal_map() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.metal_map)
  return metal_map_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_metal_map() {
  // @@protoc_insertion_point(field_release:protos.Mesh.metal_map)
  
  return metal_map_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_metal_map(::std::string* metal_map) {
  if (metal_map != NULL) {
    
  } else {
    
  }
  metal_map_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), metal_map);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.metal_map)
}

// optional string state = 11;
inline void Mesh::clear_state() {
  state_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::state() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.state)
  return state_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_state(const ::std::string& value) {
  
  state_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.state)
}
inline void Mesh::set_state(const char* value) {
  
  state_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.state)
}
inline void Mesh::set_state(const char* value, size_t size) {
  
  state_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.state)
}
inline ::std::string* Mesh::mutable_state() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.state)
  return state_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_state() {
  // @@protoc_insertion_point(field_release:protos.Mesh.state)
  
  return state_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_state(::std::string* state) {
  if (state != NULL) {
    
  } else {
    
  }
  state_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), state);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.state)
}

// optional string name = 12;
inline void Mesh::clear_name() {
  name_.ClearToEmptyNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline const ::std::string& Mesh::name() const {
  // @@protoc_insertion_point(field_get:protos.Mesh.name)
  return name_.GetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_name(const ::std::string& value) {
  
  name_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), value);
  // @@protoc_insertion_point(field_set:protos.Mesh.name)
}
inline void Mesh::set_name(const char* value) {
  
  name_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), ::std::string(value));
  // @@protoc_insertion_point(field_set_char:protos.Mesh.name)
}
inline void Mesh::set_name(const char* value, size_t size) {
  
  name_.SetNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(),
      ::std::string(reinterpret_cast<const char*>(value), size));
  // @@protoc_insertion_point(field_set_pointer:protos.Mesh.name)
}
inline ::std::string* Mesh::mutable_name() {
  
  // @@protoc_insertion_point(field_mutable:protos.Mesh.name)
  return name_.MutableNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline ::std::string* Mesh::release_name() {
  // @@protoc_insertion_point(field_release:protos.Mesh.name)
  
  return name_.ReleaseNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited());
}
inline void Mesh::set_allocated_name(::std::string* name) {
  if (name != NULL) {
    
  } else {
    
  }
  name_.SetAllocatedNoArena(&::google::protobuf::internal::GetEmptyStringAlreadyInited(), name);
  // @@protoc_insertion_point(field_set_allocated:protos.Mesh.name)
}

inline const Mesh* Mesh::internal_default_instance() {
  return &Mesh_default_instance_.get();
}
// -------------------------------------------------------------------

// Model

// repeated .protos.Mesh meshes = 1;
inline int Model::meshes_size() const {
  return meshes_.size();
}
inline void Model::clear_meshes() {
  meshes_.Clear();
}
inline const ::protos::Mesh& Model::meshes(int index) const {
  // @@protoc_insertion_point(field_get:protos.Model.meshes)
  return meshes_.Get(index);
}
inline ::protos::Mesh* Model::mutable_meshes(int index) {
  // @@protoc_insertion_point(field_mutable:protos.Model.meshes)
  return meshes_.Mutable(index);
}
inline ::protos::Mesh* Model::add_meshes() {
  // @@protoc_insertion_point(field_add:protos.Model.meshes)
  return meshes_.Add();
}
inline ::google::protobuf::RepeatedPtrField< ::protos::Mesh >*
Model::mutable_meshes() {
  // @@protoc_insertion_point(field_mutable_list:protos.Model.meshes)
  return &meshes_;
}
inline const ::google::protobuf::RepeatedPtrField< ::protos::Mesh >&
Model::meshes() const {
  // @@protoc_insertion_point(field_list:protos.Model.meshes)
  return meshes_;
}

inline const Model* Model::internal_default_instance() {
  return &Model_default_instance_.get();
}
#endif  // !PROTOBUF_INLINE_NOT_IN_HEADERS
// -------------------------------------------------------------------


// @@protoc_insertion_point(namespace_scope)

}  // namespace protos

// @@protoc_insertion_point(global_scope)

#endif  // PROTOBUF_model_2eproto__INCLUDED
