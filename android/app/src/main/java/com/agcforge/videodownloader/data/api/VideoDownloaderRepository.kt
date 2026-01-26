package com.agcforge.videodownloader.data.api

import com.agcforge.videodownloader.data.dto.AuthResponse
import com.agcforge.videodownloader.data.dto.DownloadRequest
import com.agcforge.videodownloader.data.dto.LoginRequest
import com.agcforge.videodownloader.data.model.DownloadTask
import com.agcforge.videodownloader.data.model.Platform
import com.agcforge.videodownloader.data.model.User

class VideoDownloaderRepository {

    private val api = ApiClient.apiService

    suspend fun getPlatforms(): Result<List<Platform>> {
        return try {
            val response = api.getPlatforms()
            if (response.isSuccessful && response.body()?.success == true) {
                Result.success(response.body()?.data?.platforms ?: emptyList())
            } else {
                Result.failure(Exception(response.body()?.error ?: "Unknown error"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun createDownload(url: String, platformId: String? = null): Result<DownloadTask> {
        return try {
            val request = DownloadRequest(url, platformId)
            val response = api.createDownload(request)
            if (response.isSuccessful && response.body()?.success == true) {
                response.body()?.data?.task?.let {
                    Result.success(it)
                } ?: Result.failure(Exception("No data returned"))
            } else {
                Result.failure(Exception(response.body()?.error ?: "Download failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getDownload(id: String): Result<DownloadTask> {
        return try {
            val response = api.getDownload(id)
            if (response.isSuccessful && response.body()?.success == true) {
                response.body()?.data?.let {
                    Result.success(it)
                } ?: Result.failure(Exception("No data returned"))
            } else {
                Result.failure(Exception(response.body()?.error ?: "Failed to fetch download"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getDownloads(page: Int = 1, limit: Int = 20): Result<List<DownloadTask>> {
        return try {
            val response = api.getDownloads(page, limit)
            if (response.isSuccessful && response.body()?.success == true) {
                Result.success(response.body()?.data ?: emptyList())
            } else {
                Result.failure(Exception(response.body()?.error ?: "Failed to fetch downloads"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun login(email: String, password: String): Result<AuthResponse> {
        return try {
            val request = LoginRequest(email, password)
            val response = api.login(request)
            if (response.isSuccessful && response.body()?.success == true) {
                response.body()?.data?.let {
                    ApiClient.setAuthToken(it.token)
                    Result.success(it)
                } ?: Result.failure(Exception("No data returned"))
            } else {
                Result.failure(Exception(response.body()?.error ?: "Login failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun register(request: Map<String, String>): Result<AuthResponse> {
        return try {
            val response = api.register(request)
            if (response.isSuccessful && response.body()?.success == true) {
                response.body()?.data?.let {
                    Result.success(it)
                } ?: Result.failure(Exception("No data returned"))
            } else {
                Result.failure(Exception(response.body()?.error ?: "Registration failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun forgotPassword(request: Map<String, String>): Result<Unit> {
        return try {
            val response = api.forgotPassword(request)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to send reset link"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun resetPassword(request: Map<String, String>): Result<Unit> {
        return try {
            val response = api.resetPassword(request)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to reset password"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun resendVerificationEmail(request: Map<String, String>): Result<Unit> {
        return try {
            val response = api.resendVerificationEmail(request)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Failed to resend email"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getCurrentUser(): Result<User> {
        return try {
            val response = api.getCurrentUser()
            if (response.isSuccessful && response.body()?.success == true) {
                response.body()?.data?.let {
                    Result.success(it)
                } ?: Result.failure(Exception("No data returned"))
            } else {
                Result.failure(Exception(response.body()?.error ?: "Failed to get user"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun logout(): Result<Unit> {
        return try {
            val response = api.logout()
            if (response.isSuccessful) {
                ApiClient.setAuthToken(null)
                Result.success(Unit)
            } else {
                Result.failure(Exception("Logout failed"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}