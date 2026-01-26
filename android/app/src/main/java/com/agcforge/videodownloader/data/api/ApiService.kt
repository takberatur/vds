package com.agcforge.videodownloader.data.api

import com.agcforge.videodownloader.data.dto.ApiResponse
import com.agcforge.videodownloader.data.dto.AuthResponse
import com.agcforge.videodownloader.data.dto.DownloadRequest
import com.agcforge.videodownloader.data.dto.DownloadResponse
import com.agcforge.videodownloader.data.dto.LoginRequest
import com.agcforge.videodownloader.data.dto.PlatformListResponse
import com.agcforge.videodownloader.data.model.*
import retrofit2.Response
import retrofit2.http.*

interface ApiService {

    @GET("platforms")
    suspend fun getPlatforms(): Response<ApiResponse<PlatformListResponse>>

    @GET("platforms/{id}")
    suspend fun getPlatform(@Path("id") id: String): Response<ApiResponse<Platform>>

    @POST("downloads")
    suspend fun createDownload(@Body request: DownloadRequest): Response<ApiResponse<DownloadResponse>>

    @GET("downloads/{id}")
    suspend fun getDownload(@Path("id") id: String): Response<ApiResponse<DownloadTask>>

    @GET("downloads")
    suspend fun getDownloads(
        @Query("page") page: Int = 1,
        @Query("limit") limit: Int = 20
    ): Response<ApiResponse<List<DownloadTask>>>

    @POST("auth/login")
    suspend fun login(@Body request: LoginRequest): Response<ApiResponse<AuthResponse>>

    @POST("auth/register")
    suspend fun register(@Body request: Map<String, String>): Response<ApiResponse<AuthResponse>>

    @GET("auth/me")
    suspend fun getCurrentUser(): Response<ApiResponse<User>>

    @POST("auth/logout")
    suspend fun logout(): Response<ApiResponse<Any>>
}

