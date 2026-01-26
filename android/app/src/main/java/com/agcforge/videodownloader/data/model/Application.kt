package com.agcforge.videodownloader.data.model

import android.os.Parcelable
import com.google.gson.annotations.SerializedName
import kotlinx.parcelize.Parcelize

@Parcelize
data class Application(
    @SerializedName("id") val id: String,
    @SerializedName("name") val name: String,
    @SerializedName("package_name") val packageName: String,
    @SerializedName("api_key") val apiKey: String,
    @SerializedName("secret_key") val secretKey: String,
    @SerializedName("version") val version: String? = null,
    @SerializedName("platform") val platform: String = "android",
    @SerializedName("enable_monetization") val enableMonetization: Boolean,
    @SerializedName("enable_admob") val enableAdmob: Boolean,
    @SerializedName("enable_unity_ad") val enableUnityAd: Boolean,
    @SerializedName("enable_start_app") val enableStartApp: Boolean,
    @SerializedName("enable_in_app_purchase") val enableInAppPurchase: Boolean,
    @SerializedName("is_active") val isActive: Boolean,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updatedAt: String,
    @SerializedName("in_app_products") val inAppProducts: List<InAppProduct>? = null
) : Parcelable